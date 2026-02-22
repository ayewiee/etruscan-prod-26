package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type EventTypeRepository interface {
	Create(ctx context.Context, eventType *models.EventType) (*models.EventType, error)
	GetByKey(ctx context.Context, key string) (*models.EventType, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.EventType, error)
	List(ctx context.Context) ([]*models.EventType, error)
}

type SQLCEventTypeRepository struct {
	db *dbgen.Queries
}

func NewSQLCEventTypeRepository(db *dbgen.Queries) EventTypeRepository {
	return &SQLCEventTypeRepository{db: db}
}

func (r *SQLCEventTypeRepository) Create(ctx context.Context, eventType *models.EventType) (*models.EventType, error) {
	row, err := r.db.CreateEventType(ctx, dbgen.CreateEventTypeParams{
		Key:         eventType.Key,
		Name:        eventType.Name,
		Description: database.ToPgText(eventType.Description),
		Requires:    database.ToPgUUID(eventType.RequiresID),
	})
	if err != nil {
		return nil, err
	}

	return eventTypeRowToDomain(row), nil
}

func (r *SQLCEventTypeRepository) GetByKey(ctx context.Context, key string) (*models.EventType, error) {
	row, err := r.db.GetEventTypeByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return eventTypeRowWithRequiresKeyToDomain(row.ID, row.Key, row.Name, row.Description, row.Requires, row.CreatedAt, row.RequiresKey), nil
}

func (r *SQLCEventTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EventType, error) {
	row, err := r.db.GetEventTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return eventTypeRowWithRequiresKeyToDomain(row.ID, row.Key, row.Name, row.Description, row.Requires, row.CreatedAt, row.RequiresKey), nil
}

// eventTypeRowWithRequiresKeyToDomain maps a query row (with joined requires_key) to the domain model.
func eventTypeRowWithRequiresKeyToDomain(id uuid.UUID, key, name string, desc pgtype.Text, requires pgtype.UUID, createdAt pgtype.Timestamptz, requiresKey pgtype.Text) *models.EventType {
	var t time.Time
	if ct := database.FromPgTimestamptz(createdAt); ct != nil {
		t = *ct
	}
	return &models.EventType{
		ID:          id,
		Key:         key,
		Name:        name,
		Description: database.FromPgText(desc),
		RequiresID:  database.FromPgUUID(requires),
		RequiresKey: database.FromPgText(requiresKey),
		CreatedAt:   t,
	}
}

func eventTypeRowToDomain(row dbgen.EventType) *models.EventType {
	var createdAt time.Time
	if t := database.FromPgTimestamptz(row.CreatedAt); t != nil {
		createdAt = *t
	}
	return &models.EventType{
		ID:          row.ID,
		Key:         row.Key,
		Name:        row.Name,
		Description: database.FromPgText(row.Description),
		RequiresID:  database.FromPgUUID(row.Requires),
		RequiresKey: nil,
		CreatedAt:   createdAt,
	}
}

func (r *SQLCEventTypeRepository) List(ctx context.Context) ([]*models.EventType, error) {
	rows, err := r.db.ListEventTypes(ctx)
	if err != nil {
		return nil, err
	}
	eventTypes := make([]*models.EventType, len(rows))
	for i, row := range rows {
		eventTypes[i] = eventTypeRowWithRequiresKeyToDomain(row.ID, row.Key, row.Name, row.Description, row.Requires, row.CreatedAt, row.RequiresKey)
	}
	return eventTypes, nil
}
