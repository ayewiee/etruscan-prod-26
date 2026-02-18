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

type FlagHandler struct {
	usecase *usecases.FlagUseCase
}

func NewFlagHandler(usecase *usecases.FlagUseCase) *FlagHandler {
	return &FlagHandler{usecase: usecase}
}

func modifyRequest(c echo.Context) (models.UserAuthData, *models.Flag, error) {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return models.UserAuthData{}, nil, err
	}

	var req dto.CreateUpdateFlagRequest

	if err := c.Bind(&req); err != nil {
		return models.UserAuthData{}, nil, err
	}
	if err := c.Validate(&req); err != nil {
		return models.UserAuthData{}, nil, apierrors.ValidationError(err, req)
	}

	return actor, &models.Flag{
		Key:          req.Key,
		Description:  req.Description,
		DefaultValue: req.DefaultValue,
		ValueType:    models.FlagValueType(req.ValueType),
	}, nil
}

func (h *FlagHandler) Create(c echo.Context) error {
	actor, domainFlag, err := modifyRequest(c)
	if err != nil {
		return err
	}

	flag, err := h.usecase.Create(c.Request().Context(), actor, domainFlag)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.FlagResponseFromDomain(flag))
}

func (h *FlagHandler) Update(c echo.Context) error {
	actor, domainFlag, err := modifyRequest(c)
	if err != nil {
		return err
	}

	flagId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	domainFlag.ID = flagId

	flag, err := h.usecase.Update(c.Request().Context(), actor, domainFlag)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.FlagResponseFromDomain(flag))
}

func (h *FlagHandler) List(c echo.Context) error {
	flags, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.FlagResponseListFromDomain(flags))
}

func (h *FlagHandler) GetByID(c echo.Context) error {
	flagId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	flag, err := h.usecase.GetByID(c.Request().Context(), flagId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.FlagResponseFromDomain(flag))
}

func (h *FlagHandler) Delete(c echo.Context) error {
	actor, err := api.ExtractUserAuthDataFromContext(c)
	if err != nil {
		return err
	}

	flagId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	err = h.usecase.Delete(c.Request().Context(), actor, flagId)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
