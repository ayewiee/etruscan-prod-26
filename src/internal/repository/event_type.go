package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type EventTypeRepository interface {
	GetByKey(ctx context.Context, key string) (*models.EventType, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.EventType, error)
}

type SQLCEventTypeRepository struct {
	db *dbgen.Queries
}

func NewSQLCEventTypeRepository(db *dbgen.Queries) EventTypeRepository {
	return &SQLCEventTypeRepository{db: db}
}

func (r *SQLCEventTypeRepository) GetByKey(ctx context.Context, key string) (*models.EventType, error) {
	row, err := r.db.GetEventTypeByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return eventTypeFromDB(row), nil
}

func (r *SQLCEventTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EventType, error) {
	row, err := r.db.GetEventTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return eventTypeFromDB(row), nil
}

func eventTypeFromDB(row dbgen.EventType) *models.EventType {
	return &models.EventType{
		ID:          row.ID,
		Key:         row.Key,
		Name:        row.Name,
		Description: database.FromPgText(row.Description),
		Requires:    database.FromPgUUID(row.Requires),
		CreatedAt:   database.FromPgTimestamptz(row.CreatedAt),
	}
}
