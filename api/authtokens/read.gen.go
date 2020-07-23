// Code generated by "make api"; DO NOT EDIT.
package authtokens

import (
	"context"
	"fmt"

	"github.com/hashicorp/watchtower/api"
)

func (s AuthToken) ReadAuthToken(ctx context.Context, id string) (*AuthToken, *api.Error, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("empty ID value passed into ReadAuthToken request")
	}

	if s.Client == nil {
		return nil, nil, fmt.Errorf("nil client in ReadAuthToken request")
	}

	var opts []api.Option
	if s.Scope.Id != "" {
		// If it's explicitly set here, override anything that might be in the
		// client
		opts = append(opts, api.WithScopeId(s.Scope.Id))
	}

	req, err := s.Client.NewRequest(ctx, "GET", fmt.Sprintf("%s/%s", "auth-tokens", id), nil, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating ReadAuthToken request: %w", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error performing client request during ReadAuthToken call: %w", err)
	}

	target := new(AuthToken)
	apiErr, err := resp.Decode(target)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding ReadAuthToken repsonse: %w", err)
	}

	target.Client = s.Client

	return target, apiErr, nil
}
