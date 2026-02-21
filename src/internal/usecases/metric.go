package usecases

import (
	"context"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"

	"github.com/google/uuid"
)

type MetricUseCase struct {
	repo repository.MetricRepository
}

func NewMetricUseCase(repo repository.MetricRepository) *MetricUseCase {
	return &MetricUseCase{repo: repo}
}

func (uc *MetricUseCase) Create(ctx context.Context, m *models.Metric) (*models.Metric, error) {
	return uc.repo.Create(ctx, m)
}

func (uc *MetricUseCase) GetByKey(ctx context.Context, key string) (*models.Metric, error) {
	return uc.repo.GetByKey(ctx, key)
}

func (uc *MetricUseCase) GetByID(ctx context.Context, id uuid.UUID) (*models.Metric, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *MetricUseCase) List(ctx context.Context) ([]*models.Metric, error) {
	return uc.repo.List(ctx)
}
