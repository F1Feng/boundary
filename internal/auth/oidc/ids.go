package oidc

import (
	"context"

	"github.com/hashicorp/boundary/internal/auth"
	"github.com/hashicorp/boundary/internal/db"
	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/intglobals"
	"github.com/hashicorp/boundary/internal/types/subtypes"
)

func init() {
	if err := auth.Register(Subtype, AuthMethodPrefix, AccountPrefix, intglobals.OidcManagedGroupPrefix); err != nil {
		panic(err)
	}
}

const (
	// AuthMethodPrefix defines the prefix for AuthMethod public ids.
	AuthMethodPrefix = "amoidc"
	// AccountPrefix defines the prefix for Account public ids.
	AccountPrefix = "acctoidc"

	Subtype = subtypes.Subtype("oidc")
)

func newAuthMethodId(ctx context.Context) (string, error) {
	const op = "oidc.newAuthMethodId"
	id, err := db.NewPublicId(AuthMethodPrefix)
	if err != nil {
		return "", errors.Wrap(ctx, err, op)
	}
	return id, nil
}

func newAccountId(ctx context.Context, authMethodId, issuer, sub string) (string, error) {
	const op = "oidc.newAccountId"
	if authMethodId == "" {
		return "", errors.New(ctx, errors.InvalidParameter, op, "missing auth method id")
	}
	if issuer == "" {
		return "", errors.New(ctx, errors.InvalidParameter, op, "missing issuer")
	}
	if sub == "" {
		return "", errors.New(ctx, errors.InvalidParameter, op, "missing subject")
	}
	id, err := db.NewPublicId(AccountPrefix, db.WithPrngValues([]string{authMethodId, issuer, sub}))
	if err != nil {
		return "", errors.Wrap(ctx, err, op)
	}
	return id, nil
}

func newManagedGroupId(ctx context.Context) (string, error) {
	const op = "oidc.newManagedGroupId"
	id, err := db.NewPublicId(intglobals.OidcManagedGroupPrefix)
	if err != nil {
		return "", errors.Wrap(ctx, err, op)
	}
	return id, nil
}
