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

type ApproverGroupHandler struct {
	usecase *usecases.ApproverGroupUseCase
}

func NewApproverGroupHandler(usecase *usecases.ApproverGroupUseCase) *ApproverGroupHandler {
	return &ApproverGroupHandler{usecase}
}

func (h *ApproverGroupHandler) Create(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateApproverGroupRequest

	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	ag, err := h.usecase.Create(c.Request().Context(), actor, &models.ApproverGroup{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ApproverGroupResponseFromDomain(ag))
}

func (h *ApproverGroupHandler) GetByID(c echo.Context) error {
	agId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	ag, err := h.usecase.GetByID(c.Request().Context(), agId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ApproverGroupResponseFromDomain(ag))
}

func (h *ApproverGroupHandler) AddMembers(c echo.Context) error {
	actor, agId, ids, err := h.modifyMembers(c)
	if err != nil {
		return err
	}
	ag, err := h.usecase.AddMembers(c.Request().Context(), actor, agId, ids)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ApproverGroupResponseFromDomain(ag))
}

func (h *ApproverGroupHandler) RemoveMembers(c echo.Context) error {
	actor, agId, ids, err := h.modifyMembers(c)
	if err != nil {
		return err
	}
	ag, err := h.usecase.RemoveMembers(c.Request().Context(), actor, agId, ids)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ApproverGroupResponseFromDomain(ag))
}

func (h *ApproverGroupHandler) modifyMembers(c echo.Context) (models.UserAuthData, uuid.UUID, []uuid.UUID, error) {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return models.UserAuthData{}, uuid.Nil, nil, err
	}

	agId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return models.UserAuthData{}, uuid.Nil, nil, apierrors.DumbValidationError(
			"id",
			c.Param("id"),
			"Invalid UUID",
			err,
		)
	}

	var req dto.ModifyApproverGroupMembersRequest

	if err := c.Bind(&req); err != nil {
		return models.UserAuthData{}, uuid.Nil, nil, err
	}
	if err := c.Validate(&req); err != nil {
		return models.UserAuthData{}, uuid.Nil, nil, apierrors.ValidationError(err, req)
	}

	return actor, agId, req.Users, nil
}
