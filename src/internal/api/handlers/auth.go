package handlers

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
	"etruscan/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	usecase *usecases.AuthUseCase
}

func NewAuthHandler(usecase *usecases.AuthUseCase) *AuthHandler {
	return &AuthHandler{usecase}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	token, user, err := h.usecase.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{
		Token: string(token),
		User:  dto.UserResponseDTOFromDomain(user),
	})
}
