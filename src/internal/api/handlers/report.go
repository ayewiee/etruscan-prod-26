package handlers

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/api/dto"
	"etruscan/internal/domain/models"
	"etruscan/internal/usecases"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ReportHandler struct {
	usecase *usecases.ReportUseCase
}

func NewReportHandler(usecase *usecases.ReportUseCase) *ReportHandler {
	return &ReportHandler{usecase: usecase}
}

func (h *ReportHandler) GetExperimentReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apierrors.DumbValidationError("id", c.Param("id"), "Invalid UUID", err)
	}

	var q dto.ExperimentReportQuery
	if err := c.Bind(&q); err != nil {
		return models.ErrInvalidJSON
	}
	if err := c.Validate(&q); err != nil {
		return apierrors.ValidationError(err, q)
	}

	if !q.To.After(q.From) {
		return apierrors.DumbValidationError("to", q.To, "'to' must appear (in time) 'after' from'", nil)
	}

	report, err := h.usecase.GetExperimentReport(c.Request().Context(), id, q.From, q.To)
	if err != nil {
		return err
	}

	resp := dto.ExperimentReportResponse{
		ExperimentID: report.ExperimentID,
		From:         report.From.Format(time.RFC3339),
		To:           report.To.Format(time.RFC3339),
		Variants:     make([]dto.VariantMetricValues, len(report.Variants)),
	}
	for i, v := range report.Variants {
		resp.Variants[i] = dto.VariantMetricValues{
			VariantID:   v.VariantID,
			VariantName: v.VariantName,
			Metrics:     v.Metrics,
		}
	}
	return c.JSON(http.StatusOK, resp)
}
