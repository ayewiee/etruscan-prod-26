package handlers

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
	"etruscan/internal/domain/models"
	"etruscan/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type EventsHandler struct {
	usecase *usecases.EventsUseCase
}

func NewEventsHandler(usecase *usecases.EventsUseCase) *EventsHandler {
	return &EventsHandler{usecase: usecase}
}

func (h *EventsHandler) Ingest(c echo.Context) error {
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

	result, err := h.usecase.Ingest(c.Request().Context(), items)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.BatchEventsResponseFromDomain(result))
}
