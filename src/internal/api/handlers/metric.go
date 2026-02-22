package handlers

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
	"etruscan/internal/domain/models"
	"etruscan/internal/usecases"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MetricHandler struct {
	usecase *usecases.MetricUseCase
}

func NewMetricHandler(usecase *usecases.MetricUseCase) *MetricHandler {
	return &MetricHandler{usecase: usecase}
}

func (h *MetricHandler) Create(c echo.Context) error {
	var req dto.CreateMetricRequest
	if err := c.Bind(&req); err != nil {
		return models.ErrInvalidJSON
	}
	if err := c.Validate(&req); err != nil {
		return apierrors.ValidationError(err, req)
	}

	// Either primitive (eventTypeKey + aggregationType) or derived (numeratorMetricKey + denominatorMetricKey)
	isDerived := req.NumeratorMetricKey != nil && *req.NumeratorMetricKey != "" &&
		req.DenominatorMetricKey != nil && *req.DenominatorMetricKey != ""
	if isDerived {
		if req.EventTypeKey != "" || req.AggregationType != "" {
			return apierrors.MultipleDumbValidationErrors(
				models.FieldError{Field: "eventTypeKey", Issue: "must be omitted for derived (ratio) metrics", RejectedValue: req.EventTypeKey},
				models.FieldError{Field: "aggregationType", Issue: "must be omitted for derived (ratio) metrics", RejectedValue: req.AggregationType},
			)
		}
	} else {
		if req.EventTypeKey == "" || req.AggregationType == "" {
			var errs []models.FieldError
			if req.EventTypeKey == "" {
				errs = append(errs, models.FieldError{Field: "eventTypeKey", Issue: "required for primitive metrics", RejectedValue: req.EventTypeKey})
			}
			if req.AggregationType == "" {
				errs = append(errs, models.FieldError{Field: "aggregationType", Issue: "required for primitive metrics", RejectedValue: req.AggregationType})
			}
			return apierrors.MultipleDumbValidationErrors(errs...)
		}
		if (req.NumeratorMetricKey != nil && *req.NumeratorMetricKey != "") ||
			(req.DenominatorMetricKey != nil && *req.DenominatorMetricKey != "") {
			return apierrors.MultipleDumbValidationErrors(
				models.FieldError{Field: "numeratorMetricKey", Issue: "must be omitted for primitive metrics", RejectedValue: req.NumeratorMetricKey},
				models.FieldError{Field: "denominatorMetricKey", Issue: "must be omitted for primitive metrics", RejectedValue: req.DenominatorMetricKey},
			)
		}
	}

	m := &models.Metric{
		Key:             req.Key,
		Name:            req.Name,
		Description:     req.Description,
		Type:            models.MetricType(req.Type),
		EventTypeKey:    req.EventTypeKey,
		AggregationType: models.MetricAggregationType(req.AggregationType),
		IsGuardrail:     req.IsGuardrail,
	}
	if isDerived {
		m.NumeratorMetricKey = req.NumeratorMetricKey
		m.DenominatorMetricKey = req.DenominatorMetricKey
	}

	created, err := h.usecase.Create(c.Request().Context(), m)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.MetricResponseFromDomain(created))
}

func (h *MetricHandler) List(c echo.Context) error {
	list, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return err
	}

	out := make([]*dto.MetricResponse, len(list))
	for i := range list {
		out[i] = dto.MetricResponseFromDomain(list[i])
	}

	return c.JSON(http.StatusOK, out)
}

func (h *MetricHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	m, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.MetricResponseFromDomain(m))
}
