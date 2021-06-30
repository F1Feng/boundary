// Code generated by "make handlers"; DO NOT EDIT.
package targets

import (
	"context"

	"github.com/hashicorp/boundary/internal/errors"
	pbs "github.com/hashicorp/boundary/internal/gen/controller/api/services"
	"github.com/hashicorp/boundary/internal/target"
	"github.com/hashicorp/boundary/internal/types/action"
	"github.com/hashicorp/boundary/internal/servers/controller/handlers"
)

type deleteRequest = *pbs.DeleteTargetRequest
type deleteResponse = *pbs.DeleteTargetResponse

// DeleteTarget implements the interface pbs.CredentialLibraryServiceServer.
func (s Service) DeleteTarget(ctx context.Context, req deleteRequest) (deleteResponse, error) {
	if err := validateDeleteRequest(req); err != nil {
		return nil, err
	}
	authResults := s.authResult(ctx, req.GetId(), action.Delete)
	if authResults.Error != nil {
		return nil, authResults.Error
	}
	_, err := s.deleteFromRepo(ctx, authResults.Scope.GetId(), req.GetId())
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s Service) deleteFromRepo(ctx context.Context, scopeId, id string) (bool, error) {
	const op = "targets.(Service).deleteFromRepo"
	var rows int
	var err error
	
	repo, iErr := s.repoFn()
	if iErr != nil {
		return false, iErr
	}
	rows, err = repo.DeleteTarget(ctx, scopeId, id)
	
	if err != nil {
		if errors.IsNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrap(err, op, errors.WithMsg("unable to delete resource"))
	}
	return rows > 0, nil
}

func validateDeleteRequest(req deleteRequest) error {
	return handlers.ValidateDeleteRequest(handlers.NoopValidatorFn, req, target.TcpTargetPrefix,)
}
