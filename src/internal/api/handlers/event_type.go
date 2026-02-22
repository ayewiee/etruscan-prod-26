package handlers

import (
	"etruscan/internal/api"
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
	"etruscan/internal/domain/models"
	"etruscan/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type EventTypeHandler struct {
	usecase *usecases.EventTypeUseCase
}

func NewEventTypeHandler(usecase *usecases.EventTypeUseCase) *EventTypeHandler {
	return &EventTypeHandler{usecase: usecase}
}

func (h *EventTypeHandler) Create(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateEventTypeRequest

	if err = c.Bind(&req); err != nil {
		return models.ErrInvalidJSON
	}
	if err = c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	eventType, err := h.usecase.Create(c.Request().Context(), actor, &models.EventType{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		RequiresKey: req.Requires,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.EventTypeResponseFromDomain(eventType))
}

func (h *EventTypeHandler) List(c echo.Context) error {
	eventTypes, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.EventTypeResponseListFromDomain(eventTypes))
}
