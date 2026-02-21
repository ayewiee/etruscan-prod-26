package handlers

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
	"etruscan/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type DecideHandler struct {
	usecase *usecases.DecideUseCase
}

func NewDecideHandler(usecase *usecases.DecideUseCase) *DecideHandler {
	return &DecideHandler{usecase}
}

func (h *DecideHandler) Decide(c echo.Context) error {
	var req dto.DecideRequest

	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	decision, err := h.usecase.Decide(c.Request().Context(), usecases.DecideParams{
		UserID:  req.UserID,
		FlagKey: req.FlagKey,
		Context: req.Context,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.DecisionResponseFromDomain(decision))
}
