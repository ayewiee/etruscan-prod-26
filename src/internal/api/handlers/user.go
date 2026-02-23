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

type UserHandler struct {
	usecase *usecases.UserUseCase
}

func NewUserHandler(usecase *usecases.UserUseCase) *UserHandler {
	return &UserHandler{usecase}
}

func parseUser(c echo.Context) (*models.User, *string, error) {
	var req dto.AdminCreateUpdateUserRequest

	if err := c.Bind(&req); err != nil {
		return nil, nil, err
	}
	if err := c.Validate(&req); err != nil {
		return nil, nil, apierrors.ValidationError(err, req)
	}

	return req.ToDomain(), req.Password, nil
}

func (h *UserHandler) AdminCreate(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	domainUser, password, err := parseUser(c)
	if err != nil {
		return err
	}

	if password == nil {
		return apierrors.DumbValidationError("password", nil, "Password is required", nil)
	}

	user, err := h.usecase.Create(c.Request().Context(), actor, domainUser, *password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.UserResponseDTOFromDomain(user))
}

func (h *UserHandler) AdminGetProfile(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid User ID", err)
	}

	user, err := h.usecase.GetByID(c.Request().Context(), actor, userId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.UserResponseDTOFromDomain(user))
}

func (h *UserHandler) AdminUpdate(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid User ID", err)
	}

	user, err := h.usecase.GetByID(c.Request().Context(), actor, userId)
	if err != nil {
		return err
	}

	domainUser, password, err := parseUser(c)
	if err != nil {
		return err
	}

	updatedUser, err := h.usecase.AdminUpdate(c.Request().Context(), actor, user, domainUser, password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.UserResponseDTOFromDomain(updatedUser))
}

func (h *UserHandler) AdminSoftDelete(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid User ID", err)
	}

	err = h.usecase.SoftDelete(c.Request().Context(), actor, userId)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) AdminList(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	pagination, err := api.ParsePagination(c)
	if err != nil {
		return err
	}

	users, total, err := h.usecase.List(c.Request().Context(), actor, pagination)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.PaginatedUserResponseDTOListFromDomainList(
		users,
		total,
		pagination.Page,
		pagination.Size,
	))
}

func (h *UserHandler) GetProfile(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	user, err := h.usecase.GetByID(c.Request().Context(), actor, actor.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.UserResponseDTOFromDomain(user))
}

func (h *UserHandler) UpdateProfile(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	var req dto.UserUpdateRequest

	if err = c.Bind(&req); err != nil {
		return err
	}
	if err = c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	user, err := h.usecase.Update(c.Request().Context(), actor, req.Username, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.UserResponseDTOFromDomain(user))
}
