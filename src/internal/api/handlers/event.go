package handlers

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
	"etruscan/internal/domain/models"
	"etruscan/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type EventHandler struct {
	usecase *usecases.EventUseCase
}

func NewEventHandler(usecase *usecases.EventUseCase) *EventHandler {
	return &EventHandler{usecase: usecase}
}

func (h *EventHandler) BatchTrack(c echo.Context) error {
	var req dto.BatchEventsRequest
	if err := c.Bind(&req); err != nil {
		return models.ErrInvalidJSON
	}
	if err := c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	items := make([]models.BatchEventItem, len(req.Events))
	for i := range req.Events {
		items[i] = req.Events[i].ToDomain()
	}

	result, err := h.usecase.BatchTrack(c.Request().Context(), items)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.BatchEventsResponseFromDomain(result))
}
