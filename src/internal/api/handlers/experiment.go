package handlers

import (
	"etruscan/internal/api"
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
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

func modifyRequest(c echo.Context) (models.UserAuthData, *models.Experiment, error) {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return models.UserAuthData{}, nil, err
	}

	var req dto.CreateUpdateExperimentRequest

	if err := c.Bind(&req); err != nil {
		return models.UserAuthData{}, nil, err
	}
	if err := c.Validate(&req); err != nil {
		return models.UserAuthData{}, nil, apierrors.ValidationError(err, req)
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

	return actor, &models.Experiment{
		FlagID:             req.FlagID,
		Name:               req.Name,
		Description:        req.Description,
		AudiencePercentage: req.AudiencePercentage,
		TargetingRule:      req.TargetingRule,
		Variants:           variantsDomain,
	}, nil
}

func (h *ExperimentHandler) Create(c echo.Context) error {
	actor, domainExperiment, err := modifyRequest(c)
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
	expId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	exp, err := h.usecase.GetByID(c.Request().Context(), expId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ExperimentResponseFromDomain(exp))
}

func (h *ExperimentHandler) Update(c echo.Context) error {
	expId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	actor, domainExperiment, err := modifyRequest(c)
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

func (h *ExperimentHandler) SendOnReview(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	expId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	err = h.usecase.SendOnReview(c.Request().Context(), actor, expId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully sent on review!"})
}

func (h *ExperimentHandler) review(c echo.Context, action models.ExperimentReviewAction) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	expId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
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
