package usecases

import (
	"context"
	"errors"
	"etruscan/internal/domain/models"
	"fmt"

	"github.com/google/uuid"
)

func (uc *ExperimentUseCase) SendOnReview(
	ctx context.Context,
	actor models.UserAuthData,
	id uuid.UUID,
) error {
	if !actor.Role.CanManageExperiments() {
		return models.ErrForbidden
	}

	dbExperiment, err := uc.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if actor.Role != models.UserRoleAdmin && actor.ID != dbExperiment.CreatedBy {
		return models.ErrForbidden
	}
	if !dbExperiment.Status.CanTransitionTo(models.ExperimentStatusOnReview) {
		return models.NewErrForbidden(fmt.Sprintf(
			"Experiment with status %s cannot be sent on review",
			dbExperiment.Status,
		))
	}

	err = uc.repo.ClearExperimentReviews(ctx, id)
	if err != nil {
		return err
	}

	return uc.updateStatus(ctx, dbExperiment, &models.ExperimentStatusChange{
		ExperimentID: id,
		ActorID:      &actor.ID,
		From:         &dbExperiment.Status,
		To:           models.ExperimentStatusOnReview,
	})
}

type ExperimentReviewInput struct {
	Actor   models.UserAuthData
	ID      uuid.UUID
	Action  models.ExperimentReviewAction
	Comment *string
}

func (uc *ExperimentUseCase) Review(
	ctx context.Context,
	inp ExperimentReviewInput,
) error {
	if !inp.Actor.Role.CanApprove() {
		return models.ErrForbidden
	}

	dbExperiment, err := uc.GetByID(ctx, inp.ID)
	if err != nil {
		return err
	}

	expOwner, err := uc.userRepo.GetById(ctx, dbExperiment.CreatedBy)
	if err != nil {
		return err
	}

	ok, err := uc.userCanApproveExperimentsOfThisUser(ctx, inp.Actor, expOwner)
	if err != nil {
		return err
	}
	if !ok {
		return models.ErrForbidden
	}

	if !dbExperiment.Status.ApprovalAllowed() {
		return models.NewErrForbidden(fmt.Sprintf(
			"Experiment with status %s cannot be reviewed",
			dbExperiment.Status,
		))
	}

	switch inp.Action {
	case models.ExperimentReviewActionApprove:
		return uc.approveExperiment(ctx, inp.Actor, dbExperiment)
	case models.ExperimentReviewActionRequestChanges:
		return uc.requestChangesForExperiment(ctx, dbExperiment, inp)
	case models.ExperimentReviewActionDecline:
		return uc.declineExperiment(ctx, dbExperiment, inp)
	default:
		return errors.New("illegal experiment review action")
	}
}

func (uc *ExperimentUseCase) userCanApproveExperimentsOfThisUser(
	ctx context.Context,
	actor models.UserAuthData,
	expOwner *models.User,
) (bool, error) {
	if actor.Role == models.UserRoleAdmin {
		return true, nil
	}
	// if there's no approver group assigned, only admins can approve
	if expOwner.ApproverGroup == nil {
		return false, nil
	}

	expOwnerApproverGroup, err := uc.approverGroupRepo.GetByID(ctx, *expOwner.ApproverGroup)
	if err != nil {
		return false, err
	}

	for _, member := range expOwnerApproverGroup.Members {
		if member.ID == actor.ID {
			return true, nil
		}
	}

	return false, nil
}

func (uc *ExperimentUseCase) approveExperiment(
	ctx context.Context,
	actor models.UserAuthData,
	exp *models.Experiment,
) error {
	err := uc.repo.CreateExperimentReview(ctx, &models.ExperimentReview{
		ExperimentID: exp.ID,
		ApproverID:   actor.ID,
		Decision:     models.ExperimentReviewDecisionApproved,
		Comment:      nil,
	})
	if err != nil {
		return err
	}

	hasEnoughApprovals, minApprovals, err := uc.checkApprovalsCount(ctx, exp)
	if err != nil {
		return err
	}

	if hasEnoughApprovals {
		if !exp.Status.CanTransitionTo(models.ExperimentStatusApproved) {
			return errors.New(fmt.Sprintf("experiment cannot be approved from %s", exp.Status))
		}

		systemMsg := fmt.Sprintf("Owner's minimal approval threshold reached (%d)", minApprovals)

		return uc.updateStatus(ctx, exp, &models.ExperimentStatusChange{
			ExperimentID: exp.ID,
			ActorID:      nil, // system update
			From:         &exp.Status,
			To:           models.ExperimentStatusApproved,
			Comment:      &systemMsg,
		})
	}

	return nil
}

func (uc *ExperimentUseCase) checkApprovalsCount(
	ctx context.Context,
	exp *models.Experiment,
) (bool, int, error) {
	approvalsCount, err := uc.repo.CountApprovals(ctx, exp.ID)
	if err != nil {
		return false, 0, err
	}

	user, err := uc.userRepo.GetById(ctx, exp.CreatedBy)
	if err != nil {
		return false, 0, err
	}

	var minApprovals int
	if user.MinApprovals != nil {
		minApprovals = *user.MinApprovals
	} else {
		minApprovals = uc.defaultMinApprovals
	}

	if approvalsCount >= minApprovals {
		return true, minApprovals, nil
	}

	return false, minApprovals, nil
}

func (uc *ExperimentUseCase) requestChangesForExperiment(
	ctx context.Context,
	exp *models.Experiment,
	inp ExperimentReviewInput,
) error {
	if !exp.Status.CanTransitionTo(models.ExperimentStatusDraft) {
		return errors.New(fmt.Sprintf(
			"changes cannot be requested for experiment with status %s",
			exp.Status,
		))
	}

	err := uc.repo.CreateExperimentReview(ctx, &models.ExperimentReview{
		ExperimentID: exp.ID,
		ApproverID:   inp.Actor.ID,
		Decision:     models.ExperimentReviewDecisionChangesRequested,
		Comment:      inp.Comment,
	})
	if err != nil {
		return err
	}

	systemMsg := "Approver requested changes"

	return uc.updateStatus(ctx, exp, &models.ExperimentStatusChange{
		ExperimentID: exp.ID,
		ActorID:      nil, // system update
		From:         &exp.Status,
		To:           models.ExperimentStatusDraft,
		Comment:      &systemMsg,
	})

}

func (uc *ExperimentUseCase) declineExperiment(
	ctx context.Context,
	exp *models.Experiment,
	inp ExperimentReviewInput,
) error {
	if !exp.Status.CanTransitionTo(models.ExperimentStatusDeclined) {
		return errors.New(fmt.Sprintf(
			"changes cannot be requested for experiment with status %s",
			exp.Status,
		))
	}

	err := uc.repo.CreateExperimentReview(ctx, &models.ExperimentReview{
		ExperimentID: exp.ID,
		ApproverID:   inp.Actor.ID,
		Decision:     models.ExperimentReviewDecisionDeclined,
		Comment:      inp.Comment,
	})
	if err != nil {
		return err
	}

	systemMsg := "Approver declined the experiment"

	return uc.updateStatus(ctx, exp, &models.ExperimentStatusChange{
		ExperimentID: exp.ID,
		ActorID:      nil, // system update
		From:         &exp.Status,
		To:           models.ExperimentStatusDeclined,
		Comment:      &systemMsg,
	})

	// TODO: notify about this
}
