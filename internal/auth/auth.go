package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hashicorp/boundary/globals"
	"github.com/hashicorp/boundary/internal/gen/controller/api/resources/scopes"
	"github.com/hashicorp/boundary/internal/kms"
	"github.com/hashicorp/boundary/internal/perms"
	"github.com/hashicorp/boundary/internal/servers/controller/common"
	"github.com/hashicorp/boundary/internal/servers/controller/handlers"
	"github.com/hashicorp/boundary/internal/types/action"
	"github.com/hashicorp/boundary/internal/types/resource"
	"github.com/hashicorp/boundary/internal/types/scope"
	"github.com/hashicorp/boundary/recovery"
	"github.com/hashicorp/go-hclog"
	wrapping "github.com/hashicorp/go-kms-wrapping"
	"github.com/kr/pretty"
	"google.golang.org/protobuf/proto"
)

const (
	HeaderAuthMethod    = "Authorization"
	HttpOnlyCookieName  = "wt-http-token-cookie"
	JsVisibleCookieName = "wt-js-token-cookie"
)

type TokenFormat int

const (
	AuthTokenTypeUnknown TokenFormat = iota

	// Came in via the Authentication: Bearer header
	AuthTokenTypeBearer

	// Came in via split cookies
	AuthTokenTypeSplitCookie

	// It's of recovery type
	AuthTokenTypeRecoveryKms

	// It's _known_ to be invalid, that is, a token was provided that is simply
	// not valid and may be part of a DDoS or other attack
	AuthTokenTypeInvalid
)

type key int

var verifierKey key

// RequestInfo contains request parameters necessary for checking authn/authz
type RequestInfo struct {
	Path        string
	Method      string
	PublicId    string
	Token       string
	TokenFormat TokenFormat

	// This is used for operations on the scopes collection
	scopeIdOverride string
	userIdOverride  string

	// The following are useful for tests
	DisableAuthzFailures bool
	DisableAuthEntirely  bool
}

type VerifyResults struct {
	UserId string
	Error  error
	Scope  *scopes.ScopeInfo
}

type verifier struct {
	logger          hclog.Logger
	iamRepoFn       common.IamRepoFactory
	authTokenRepoFn common.AuthTokenRepoFactory
	serversRepoFn   common.ServersRepoFactory
	kms             *kms.Kms
	requestInfo     RequestInfo
	res             *perms.Resource
	act             action.Type
	ctx             context.Context
}

// NewVerifierContext creates a context that carries a verifier object from the
// HTTP handlers to the gRPC service handlers. It should only be created in the
// HTTP handler and should exist for every request that reaches the service
// handlers.
func NewVerifierContext(ctx context.Context,
	logger hclog.Logger,
	iamRepoFn common.IamRepoFactory,
	authTokenRepoFn common.AuthTokenRepoFactory,
	serversRepoFn common.ServersRepoFactory,
	kms *kms.Kms,
	requestInfo RequestInfo) context.Context {
	return context.WithValue(ctx, verifierKey, &verifier{
		logger:          logger,
		iamRepoFn:       iamRepoFn,
		authTokenRepoFn: authTokenRepoFn,
		serversRepoFn:   serversRepoFn,
		kms:             kms,
		requestInfo:     requestInfo,
	})
}

// Verify takes in a context that has expected parameters as values and runs an
// authn/authz check. It returns a user ID, the scope ID for the request (which
// may come from the URL and may come from the token) and whether or not to
// proceed, e.g. whether the authn/authz check resulted in failure. If an error
// occurs it's logged to the system log.
func Verify(ctx context.Context, opt ...Option) (ret VerifyResults) {
	ret.Error = handlers.ForbiddenError()
	v, ok := ctx.Value(verifierKey).(*verifier)
	if !ok {
		// We don't have a logger yet and this should never happen in any
		// context we won't catch in tests
		panic("no verifier information found in context")
	}
	opts := getOpts(opt...)
	ret.Scope = new(scopes.ScopeInfo)
	if v.requestInfo.DisableAuthEntirely {
		ret.Scope.Id = v.requestInfo.scopeIdOverride
		switch {
		case ret.Scope.Id == "global":
			ret.Scope.Type = "global"
		case strings.HasPrefix(ret.Scope.Id, scope.Org.Prefix()):
			ret.Scope.Type = scope.Org.String()
		case strings.HasPrefix(ret.Scope.Id, scope.Project.Prefix()):
			ret.Scope.Type = scope.Project.String()
		}
		ret.UserId = v.requestInfo.userIdOverride
		ret.Error = nil
		return
	}

	// table stakes
	if v.requestInfo.TokenFormat == AuthTokenTypeInvalid {
		v.logger.Trace("got invalid token type in auth function, which should not have occurred")
		return
	}

	v.ctx = ctx
	v.requestInfo.scopeIdOverride = opts.withScopeId
	if err := v.parseAuthParams(); err != nil {
		v.logger.Trace("error reading auth parameters from URL", "url", v.requestInfo.Path, "method", v.requestInfo.Method, "error", err)
		return
	}
	if v.res == nil {
		v.logger.Trace("got nil resource information after decorating auth parameters")
		return
	}

	var authResults *perms.ACLResults
	var err error
	authResults, ret.UserId, ret.Scope, err = v.performAuthCheck()
	if err != nil {
		v.logger.Error("error performing authn/authz check", "error", err)
		return
	}
	if !authResults.Allowed {
		if v.requestInfo.DisableAuthzFailures {
			ret.Error = nil
			// TODO: Decide whether to remove this
			v.logger.Info("failed authz info for request", "resource", pretty.Sprint(v.res), "user_id", ret.UserId, "action", v.act.String())
		} else {
			return
		}
	}

	ret.Error = nil
	return
}

func (v *verifier) parseAuthParams() error {
	// Remove trailing and leading slashes
	trimmedPath := strings.Trim(v.requestInfo.Path, "/")
	// Remove `v1/`
	splitPath := strings.Split(strings.TrimPrefix(trimmedPath, "v1/"), "/")
	splitLen := len(splitPath)

	// It must be at least length 1 and the first segment must be "scopes"
	switch {
	case splitLen == 0:
		return fmt.Errorf("parse auth params: invalid path")
	case splitPath[0] != "scopes":
		return fmt.Errorf("parse auth params: invalid first segment %q", splitPath[0])
	}

	for i := 1; i < splitLen; i++ {
		if splitPath[i] == "" {
			return fmt.Errorf("parse auth params: empty segment found")
		}
	}

	v.act = action.Unknown
	v.res = &perms.Resource{
		// Start out with scope, and replace when we walk backwards if it's
		// actually something else
		Type: resource.Scope,
	}

	// Handle non-custom types. We'll deal with custom types, including list,
	// after parsing the path.
	switch v.requestInfo.Method {
	case "GET":
		v.act = action.Read
	case "POST":
		v.act = action.Create
	case "PATCH":
		v.act = action.Update
	case "DELETE":
		v.act = action.Delete
	default:
		return fmt.Errorf("parse auth params: unknown method %q", v.requestInfo.Method)
	}

	// Look for a custom action
	colonSplit := strings.Split(splitPath[splitLen-1], ":")
	switch len(colonSplit) {
	case 1:
		// No custom action specified

	case 2:
		// Parse and validate the action, then elide it
		actStr := colonSplit[len(colonSplit)-1]
		v.act = action.Map[actStr]
		if v.act == action.Unknown || v.act == action.All {
			return fmt.Errorf("parse auth params: unknown action %q", actStr)
		}
		// Keep going with the logic without the custom action
		splitPath[splitLen-1] = colonSplit[0]

	default:
		return fmt.Errorf("parse auth params: unexpected number of colons in last segment %q", colonSplit[len(colonSplit)-1])
	}

	// Get scope information and handle it in a special case; that is, for
	// operating on scopes, scope from the request ID, not the path scope
	switch splitLen {
	case 1:
		// We've already validated that this is "scopes"
		if v.act == action.Read {
			v.act = action.List
		}
		if v.requestInfo.scopeIdOverride == "" {
			return errors.New("parse auth params: missing scope ID information for scopes collection operation")
		}
		v.res.ScopeId = v.requestInfo.scopeIdOverride
		return nil

	case 2:
		id := splitPath[1]
		// The next segment should be the scope ID, but it takes place not in
		// its own scope but in the parent scope. Rather than require the user
		// to provide it, look up the parent.
		switch {
		case id == "global", strings.HasPrefix(id, scope.Org.Prefix()):
			// Org scope parent is always global. Set scope for global
			// operations to global as well (it's basically acting as its own
			// parent scope). We want this so that users can e.g. modify the
			// name or description of the global scope if they have permissions
			// in the scope.
			v.res.ScopeId = "global"

		default:
			// Project case
			iamRepo, err := v.iamRepoFn()
			if err != nil {
				return fmt.Errorf("perform auth check: failed to get iam repo: %w", err)
			}

			scp, err := iamRepo.LookupScope(v.ctx, id)
			if err != nil {
				return fmt.Errorf("perform auth check: failed to lookup scope: %w", err)
			}
			if scp == nil {
				return fmt.Errorf("perform auth check: non-existent scope %q", id)
			}
			v.res.ScopeId = scp.GetParentId()
		}
		v.res.Id = id
		return nil

	case 3:
		// If a custom action was being performed within a scope, it will have
		// been elided above. If the last path segment is now empty, address
		// this scenario. In this case the action took place _in_ the scope so
		// it should be bound accordingly. (This is for actions like
		// /scopes/o_abc/:deauthenticate where the action is _in_ the scope, not
		// on the scope.)
		if splitPath[2] == "" {
			v.res.ScopeId = splitPath[1]
			v.res.Id = splitPath[1]
			return nil
		}

		fallthrough

	default:
		// In all other cases the scope ID is the next segment
		v.res.ScopeId = splitPath[1]
	}

	// Walk backwards. As we walk backwards we look for a pin and figure out if
	// we're operating on a resource or a collection. The rules for the pin are
	// as follows:
	//
	// * If the last segment is a collection, the pin is the immediately
	// preceding ID. This does not include scopes since those are permission
	// boundaries.
	//
	// * If the last segment is an ID, the pin is the immediately preceding ID
	// not including the last segment
	//
	// * If at the end of the logic the pin is the id of a scope then there is
	// no pin. The scopes are already enclosing so a pin is redundant.
	nextIdIsPin := true
	// Use an empty string so we can detect if we found anything in this loop.
	var foundId string
	var typStr string
	// We stop at [2] because we've already dealt with the first two segments
	// (scopes/<scope_id> above.
	for i := splitLen - 1; i >= 2; i-- {
		segment := splitPath[i]

		if segment == "" {
			return fmt.Errorf("parse auth parameters: unexpected empty segment")
		}

		// If the segment contains whitespace, it's not valid
		if fields := strings.Fields(segment); len(fields) != 1 || fields[0] != segment {
			return fmt.Errorf("parse auth params: segment %q contains whitespace", segment)
		}

		// Collections don't contain underscores; every resource ID does.
		segmentIsCollection := !strings.Contains(segment, "_")

		// If we see an ID, ensure that it's not the right-most ID; if not, it's
		// the pin
		if !segmentIsCollection && i != splitLen-1 && nextIdIsPin {
			v.res.Pin = segment
			// By definition this is the last thing we'd be looking for as
			// scopes were found above, so we can now break out
			break
		}

		if typStr == "" {
			// The resource check takes place inside the type check because if
			// we've identified the type we have either already identified the
			// right-most resource ID or we're operating on a collection, so
			// this prevents us from finding a different ID earlier in the path.
			// We still work backwards to identify a pin.
			if foundId == "" && !segmentIsCollection {
				foundId = segment
			} else {
				// Every collection is the plural of the resource type so drop
				// the last 's'
				if !strings.HasSuffix(segment, "s") {
					return fmt.Errorf("parse auth params: invalid collection syntax for %q", segment)
				}
				typStr = strings.TrimSuffix(segment, "s")
			}
		}
	}

	if foundId != "" {
		v.res.Id = foundId
	}

	if typStr != "" {
		v.res.Type = resource.Map[typStr]
		if v.res.Type == resource.Unknown {
			return fmt.Errorf("parse auth params: unknown resource type %q", typStr)
		}
	} else {
		// If we found no other type information, we walked backwards all the
		// way to the scope boundary, so the type is scope
		v.res.Type = resource.Scope
	}

	// If we're operating on a collection (that is, the ID is blank) and it's a
	// GET, it's actually a list
	if v.res.Id == "" && v.act == action.Read {
		v.act = action.List
	}

	// If the pin ended up being a scope, nil it out
	if v.res.Pin != "" && v.res.Pin == v.res.ScopeId {
		v.res.Pin = ""
	}

	return nil
}

func (v verifier) performAuthCheck() (aclResults *perms.ACLResults, userId string, scopeInfo *scopes.ScopeInfo, retErr error) {
	// Ensure we return an error by default if we forget to set this somewhere
	retErr = errors.New("unknown")
	// Make the linter happy
	_ = retErr
	scopeInfo = new(scopes.ScopeInfo)
	userId = "u_anon"

	// Validate the token and fetch the corresponding user ID
	switch v.requestInfo.TokenFormat {
	case AuthTokenTypeBearer, AuthTokenTypeSplitCookie:
		if v.requestInfo.Token == "" {
			// This will end up staying as the anonymous user
			break
		}
		tokenRepo, err := v.authTokenRepoFn()
		if err != nil {
			retErr = fmt.Errorf("perform auth check: failed to get authtoken repo: %w", err)
			return
		}
		at, err := tokenRepo.ValidateToken(v.ctx, v.requestInfo.PublicId, v.requestInfo.Token)
		if err != nil {
			retErr = fmt.Errorf("perform auth check: failed to validate token: %w", err)
			return
		}
		if at != nil {
			userId = at.GetIamUserId()
		}

	case AuthTokenTypeRecoveryKms:
		userId = "u_recovery"
		if v.kms == nil {
			retErr = errors.New("perform auth check: no KMS object available to authz system")
			return
		}
		wrapper := v.kms.GetExternalWrappers().Recovery()
		if wrapper == nil {
			retErr = errors.New("perform auth check: no recovery KMS is available")
			return
		}
		info, err := recovery.ParseRecoveryToken(v.ctx, wrapper, v.requestInfo.Token)
		if err != nil {
			retErr = fmt.Errorf("perform auth check: error validating recovery token: %w", err)
			return
		}
		// If we add the validity period to the creation time (which we've
		// verified is before the current time, with a minute of fudging), and
		// it's before now, it's expired and might be a replay.
		if info.CreationTime.Add(globals.RecoveryTokenValidityPeriod).Before(time.Now()) {
			retErr = errors.New("WARNING: perform auth check: recovery token has expired (possible replay attack)")
			return
		}
		repo, err := v.serversRepoFn()
		if err != nil {
			retErr = fmt.Errorf("perform auth check: error fetching servers repo: %w", err)
			return
		}
		if err := repo.AddRecoveryNonce(v.ctx, info.Nonce); err != nil {
			retErr = fmt.Errorf("WARNING: perform auth check: error adding nonce to database (possible replay attack): %w", err)
			return
		}
		v.logger.Warn("NOTE: recovery KMS was used to authorize a call", "url", v.requestInfo.Path, "method", v.requestInfo.Method)
	}

	iamRepo, err := v.iamRepoFn()
	if err != nil {
		retErr = fmt.Errorf("perform auth check: failed to get iam repo: %w", err)
		return
	}

	// Look up scope details to return. We can skip a lookup when using the
	// global scope
	switch v.res.ScopeId {
	case "global":
		scopeInfo = &scopes.ScopeInfo{
			Id:            scope.Global.String(),
			Type:          scope.Global.String(),
			Name:          scope.Global.String(),
			Description:   "Global Scope",
			ParentScopeId: "",
		}

	default:
		scp, err := iamRepo.LookupScope(v.ctx, v.res.ScopeId)
		if err != nil {
			retErr = fmt.Errorf("perform auth check: failed to lookup scope: %w", err)
			return
		}
		if scp == nil {
			retErr = fmt.Errorf("perform auth check: non-existent scope %q", v.res.ScopeId)
			return
		}
		scopeInfo = &scopes.ScopeInfo{
			Id:            scp.GetPublicId(),
			Type:          scp.GetType(),
			Name:          scp.GetName(),
			Description:   scp.GetDescription(),
			ParentScopeId: scp.GetParentId(),
		}
	}

	// At this point we don't need to look up grants since it's automatically allowed
	if v.requestInfo.TokenFormat == AuthTokenTypeRecoveryKms {
		aclResults = &perms.ACLResults{Allowed: true}
		retErr = nil
		return
	}

	var parsedGrants []perms.Grant
	var grantPairs []perms.GrantPair

	// Fetch and parse grants for this user ID (which may include grants for
	// u_anon and u_auth)
	grantPairs, err = iamRepo.GrantsForUser(v.ctx, userId)
	if err != nil {
		retErr = fmt.Errorf("perform auth check: failed to query for user grants: %w", err)
		return
	}
	parsedGrants = make([]perms.Grant, 0, len(grantPairs))
	for _, pair := range grantPairs {
		parsed, err := perms.Parse(pair.ScopeId, userId, pair.Grant)
		if err != nil {
			retErr = fmt.Errorf("perform auth check: failed to parse grant %#v: %w", pair.Grant, err)
			return
		}
		parsedGrants = append(parsedGrants, parsed)
	}

	acl := perms.NewACL(parsedGrants...)
	allowed := acl.Allowed(*v.res, v.act)

	aclResults = &allowed
	retErr = nil
	return
}

// GetTokenFromRequest pulls the token from either the Authorization header or
// split cookies and parses it. If it cannot be parsed successfully, the issue
// is logged and we return blank, so logic will continue as the anonymous user.
// The public ID and token are returned along with the token format.
func GetTokenFromRequest(logger hclog.Logger, kmsCache *kms.Kms, req *http.Request) (string, string, TokenFormat) {
	// First, get the token, either from the authorization header or from split
	// cookies
	var receivedTokenType TokenFormat
	var fullToken string
	if authHeader := req.Header.Get("Authorization"); authHeader != "" {
		headerSplit := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
		if len(headerSplit) == 2 && strings.EqualFold(strings.TrimSpace(headerSplit[0]), "bearer") {
			receivedTokenType = AuthTokenTypeBearer
			fullToken = strings.TrimSpace(headerSplit[1])
		}
	}
	if receivedTokenType != AuthTokenTypeBearer {
		var httpCookiePayload string
		var jsCookiePayload string
		if hc, err := req.Cookie(HttpOnlyCookieName); err == nil {
			httpCookiePayload = hc.Value
		}
		if jc, err := req.Cookie(JsVisibleCookieName); err == nil {
			jsCookiePayload = jc.Value
		}
		if httpCookiePayload != "" && jsCookiePayload != "" {
			receivedTokenType = AuthTokenTypeSplitCookie
			fullToken = jsCookiePayload + httpCookiePayload
		}
	}

	if receivedTokenType == AuthTokenTypeUnknown || fullToken == "" {
		// We didn't find auth info or a client screwed up and put in a blank
		// header instead of nothing at all, so return blank which will indicate
		// the anonymouse user
		return "", "", AuthTokenTypeUnknown
	}

	if strings.HasPrefix(fullToken, "r_") {
		return "", fullToken, AuthTokenTypeRecoveryKms
	}

	splitFullToken := strings.Split(fullToken, "_")
	if len(splitFullToken) != 3 {
		logger.Trace("get token from request: unexpected number of segments in token", "expected", 3, "found", len(splitFullToken))
		return "", "", AuthTokenTypeInvalid
	}

	publicId := strings.Join(splitFullToken[0:2], "_")

	// This is hardcoded to global because at this point we don't know the
	// scope. But that's okay; this is not really a security feature so much as
	// an anti-DDoSing-the-backing-database feature.
	tokenWrapper, err := kmsCache.GetWrapper(req.Context(), scope.Global.String(), kms.KeyPurposeTokens)
	if err != nil {
		logger.Warn("get token from request: unable to get wrapper for tokens", "error", err)
		return "", "", AuthTokenTypeInvalid
	}

	version := string(splitFullToken[2][0])
	switch version {
	case globals.TokenEncryptionVersion:
	default:
		logger.Trace("unknown token encryption version", "version", version)
		return "", "", AuthTokenTypeInvalid
	}
	marshaledToken := base58.Decode(splitFullToken[2][1:])

	blobInfo := new(wrapping.EncryptedBlobInfo)
	if err := proto.Unmarshal(marshaledToken, blobInfo); err != nil {
		logger.Trace("error decoding encrypted token", "error", err)
		return "", "", AuthTokenTypeInvalid
	}

	tokenBytes, err := tokenWrapper.Decrypt(req.Context(), blobInfo, []byte(publicId))
	if err != nil {
		logger.Trace("error decrypting encrypted token", "error", err)
		return "", "", AuthTokenTypeInvalid
	}
	token := string(tokenBytes)

	if receivedTokenType == AuthTokenTypeUnknown || token == "" || publicId == "" {
		logger.Trace("get token from request: after parsing, could not find valid token")
		return "", "", AuthTokenTypeUnknown
	}

	return publicId, token, receivedTokenType
}
