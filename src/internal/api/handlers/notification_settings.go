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

type NotificationSettingsHandler struct {
	usecase *usecases.NotificationSettingsUseCase
}

func NewNotificationSettingsHandler(usecase *usecases.NotificationSettingsUseCase) *NotificationSettingsHandler {
	return &NotificationSettingsHandler{usecase}
}

func (h *NotificationSettingsHandler) Create(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	expID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	var req dto.CreateNotificationSettings

	if err = c.Bind(&req); err != nil {
		return models.ErrInvalidJSON
	}
	if err = c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	err = h.usecase.Create(c.Request().Context(), &models.NotificationSettings{
		ID:             uuid.UUID{},
		UserID:         actor.ID,
		ExperimentID:   expID,
		Severity:       models.NotificationSeverity(req.Severity),
		EnableTelegram: req.EnableTelegram,
		EnableEmail:    req.EnableEmail,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "Successfully created!"})
}

func (h *NotificationSettingsHandler) DeleteForExperiment(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	expID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	err = h.usecase.DeleteForExperimentAndUser(c.Request().Context(), actor, expID)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
