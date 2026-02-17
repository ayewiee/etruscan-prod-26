package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/api/apierrors"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"

	"github.com/google/uuid"
)

type ApproverGroupUseCase struct {
	repo     repository.ApproverGroupRepository
	userRepo repository.UserRepository
}

func NewApproverGroupUseCase(
	repo repository.ApproverGroupRepository,
	userRepo repository.UserRepository,
) *ApproverGroupUseCase {
	return &ApproverGroupUseCase{repo: repo, userRepo: userRepo}
}

func (uc *ApproverGroupUseCase) Create(
	ctx context.Context,
	actor models.UserAuthData,
	approverGroup models.ApproverGroup,
) (models.ApproverGroup, error) {
	if actor.Role != models.UserRoleAdmin {
		return models.ApproverGroup{}, models.ErrForbidden
	}

	return uc.repo.Create(ctx, approverGroup)
}

func (uc *ApproverGroupUseCase) GetByID(ctx context.Context, id uuid.UUID) (models.ApproverGroup, error) {
	ag, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ApproverGroup{}, models.NewErrNotFound("Approver group not found", nil, err)
		}
	}
	return ag, nil
}

func (uc *ApproverGroupUseCase) AddMembers(
	ctx context.Context,
	actor models.UserAuthData,
	id uuid.UUID,
	members []uuid.UUID,
) (models.ApproverGroup, error) {
	if actor.Role != models.UserRoleAdmin {
		return models.ApproverGroup{}, models.ErrForbidden
	}

	if err := uc.checkUsersExistence(ctx, members); err != nil {
		return models.ApproverGroup{}, err
	}

	err := uc.repo.AddMembers(ctx, id, members)
	if err != nil {
		return models.ApproverGroup{}, err
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *ApproverGroupUseCase) RemoveMembers(
	ctx context.Context,
	actor models.UserAuthData,
	id uuid.UUID,
	members []uuid.UUID,
) (models.ApproverGroup, error) {
	if actor.Role != models.UserRoleAdmin {
		return models.ApproverGroup{}, models.ErrForbidden
	}

	if err := uc.checkUsersExistence(ctx, members); err != nil {
		return models.ApproverGroup{}, err
	}

	err := uc.repo.RemoveMembers(ctx, id, members)
	if err != nil {
		return models.ApproverGroup{}, err
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *ApproverGroupUseCase) checkUsersExistence(
	ctx context.Context,
	members []uuid.UUID,
) error {
	allExist, err := uc.userRepo.ValidateApproversExistenceAndRole(ctx, members)
	if err != nil {
		return err
	}

	if !allExist {
		return apierrors.DumbValidationError(
			"users",
			"UUIDs",
			"invalid users: some do not exist or have too low privileges",
			nil,
		)
	}

	return nil
}
