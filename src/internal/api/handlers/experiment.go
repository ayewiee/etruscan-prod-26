package handlers

import (
	"etruscan/internal/api"
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
	"etruscan/internal/common/pagination"
	"etruscan/internal/domain/models"
	"etruscan/internal/usecases"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ExperimentHandler struct {
	usecase *usecases.ExperimentUseCase
}

func NewExperimentHandler(usecase *usecases.ExperimentUseCase) *ExperimentHandler {
	return &ExperimentHandler{usecase}
}

func (h *ExperimentHandler) parseActorAndEntityId(c echo.Context) (models.UserAuthData, uuid.UUID, error) {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return models.UserAuthData{}, uuid.Nil, err
	}

	expId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return models.UserAuthData{}, uuid.Nil, apierrors.DumbValidationError(
			"id",
			c.Param("id"),
			"Invalid UUID",
			err,
		)
	}

	return actor, expId, nil
}

func parseEntityModification(c echo.Context) (*models.Experiment, error) {
	var req dto.CreateUpdateExperimentRequest

	if err := c.Bind(&req); err != nil {
		return nil, models.ErrInvalidJSON
	}
	if err := c.Validate(&req); err != nil {
		return nil, apierrors.ValidationError(err, req)
	}

	variantsDomain := make([]*models.Variant, len(req.Variants))
	for i, v := range req.Variants {
		var isControl bool
		if v.IsControl != nil {
			isControl = *v.IsControl
		} else {
			isControl = false
		}
		variantsDomain[i] = &models.Variant{
			Name:      v.Name,
			Value:     v.Value,
			Weight:    v.Weight,
			IsControl: isControl,
		}
	}

	var guardrails []*models.Guardrail
	if len(req.Guardrails) > 0 {
		guardrails = make([]*models.Guardrail, len(req.Guardrails))
		for i, gd := range req.Guardrails {
			guardrails[i] = &models.Guardrail{
				MetricKey:          gd.MetricKey,
				Threshold:          gd.Threshold,
				ThresholdDirection: gd.ThresholdDirection,
				WindowSeconds:      gd.WindowSeconds,
				Action:             gd.Action,
			}
		}
	}

	return &models.Experiment{
		FlagID:             req.FlagID,
		Name:               req.Name,
		Description:        req.Description,
		AudiencePercentage: req.AudiencePercentage,
		TargetingRule:      req.TargetingRule,
		Variants:           variantsDomain,
		MetricKeys:         req.MetricKeys,
		PrimaryMetricKey:   req.PrimaryMetricKey,
		Guardrails:         guardrails,
	}, nil
}

func (h *ExperimentHandler) Create(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	domainExperiment, err := parseEntityModification(c)
	if err != nil {
		return err
	}

	exp, err := h.usecase.Create(c.Request().Context(), actor, domainExperiment)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ExperimentResponseFromDomain(exp))
}

func (h *ExperimentHandler) GetByID(c echo.Context) error {
	_, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	exp, err := h.usecase.GetByID(c.Request().Context(), expId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ExperimentResponseFromDomain(exp))
}

func (h *ExperimentHandler) List(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	var req dto.ExperimentListFiltersQuery

	if err = c.Bind(&req); err != nil {
		return models.ErrInvalidJSON
	}
	if err = c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	pg := pagination.NewPagination(req.Page, req.Size)

	experiments, total, err := h.usecase.List(c.Request().Context(), actor, usecases.ExperimentListFilters{
		FlagID:     req.FlagID,
		CreatedBy:  req.CreatedBy,
		Status:     (*models.ExperimentStatus)(req.Status),
		Outcome:    (*models.ExperimentOutcome)(req.Outcome),
		Pagination: pg,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Items: dto.ExperimentResponseListFromDomain(experiments),
		Total: total,
		Page:  pg.Page,
		Size:  pg.Size,
	})
}

func (h *ExperimentHandler) ListStatusChanges(c echo.Context) error {
	_, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	changes, err := h.usecase.ListStatusChanges(c.Request().Context(), expId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ExperimentStatusChangeResponseListFromDomain(changes))
}

func (h *ExperimentHandler) Update(c echo.Context) error {
	actor, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	domainExperiment, err := parseEntityModification(c)
	if err != nil {
		return err
	}

	domainExperiment.ID = expId

	exp, err := h.usecase.Update(c.Request().Context(), actor, domainExperiment)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ExperimentResponseFromDomain(exp))
}

func (h *ExperimentHandler) ListSnapshots(c echo.Context) error {
	_, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	snapshots, err := h.usecase.ListSnapshots(c.Request().Context(), expId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ExperimentSnapshotResponseListFromDomain(snapshots))
}

func (h *ExperimentHandler) SendOnReview(c echo.Context) error {
	actor, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	err = h.usecase.SendOnReview(c.Request().Context(), actor, expId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully sent on review!"})
}

func (h *ExperimentHandler) review(c echo.Context, action models.ExperimentReviewAction) error {
	actor, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	var comment *string

	if action != models.ExperimentReviewActionApprove {
		var req dto.ExperimentReviewRequest

		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(&req); err != nil {
			return apierrors.ValidationError(err, req)
		}

		comment = &req.Comment
	}

	err = h.usecase.Review(c.Request().Context(), usecases.ExperimentReviewInput{
		Actor:   actor,
		ID:      expId,
		Action:  action,
		Comment: comment,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully reviewed!"})
}

func (h *ExperimentHandler) Approve(c echo.Context) error {
	return h.review(c, models.ExperimentReviewActionApprove)
}

func (h *ExperimentHandler) RequestChanges(c echo.Context) error {
	return h.review(c, models.ExperimentReviewActionRequestChanges)
}

func (h *ExperimentHandler) Decline(c echo.Context) error {
	return h.review(c, models.ExperimentReviewActionDecline)
}

func (h *ExperimentHandler) Launch(c echo.Context) error {
	actor, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	err = h.usecase.Launch(c.Request().Context(), actor, expId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully launched!"})
}

func (h *ExperimentHandler) Pause(c echo.Context) error {
	actor, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	err = h.usecase.Pause(c.Request().Context(), actor, expId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully paused!"})
}

func (h *ExperimentHandler) Finish(c echo.Context) error {
	actor, expId, err := h.parseActorAndEntityId(c)
	if err != nil {
		return err
	}

	var req dto.FinishExperimentRequest
	if err := c.Bind(&req); err != nil {
		return models.ErrInvalidJSON
	}
	if err := c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	outcome := models.ExperimentOutcome(req.Outcome)
	exp, err := h.usecase.Finish(c.Request().Context(), actor, expId, outcome, req.Comment)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ExperimentResponseFromDomain(exp))
}
