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

type EventRepository interface {
	Create(ctx context.Context, params CreateEventParams) (uuid.UUID, error)
	ExistsByTypeAndClientID(ctx context.Context, eventTypeKey, clientEventID string) (bool, error)
	ListByDecisionIDsAndWindow(ctx context.Context, decisionIDs []uuid.UUID, from, to time.Time) ([]*models.Event, error)
}

type CreateEventParams struct {
	EventTypeKey  string
	DecisionID    *uuid.UUID
	UserID        string
	Properties    []byte
	Timestamp     pgtype.Timestamptz
	ClientEventID string
}

type SQLCEventRepository struct {
	db *dbgen.Queries
}

func NewSQLCEventRepository(db *dbgen.Queries) EventRepository {
	return &SQLCEventRepository{db: db}
}

func (r *SQLCEventRepository) Create(ctx context.Context, params CreateEventParams) (uuid.UUID, error) {
	var clientID pgtype.Text
	if params.ClientEventID != "" {
		clientID = pgtype.Text{String: params.ClientEventID, Valid: true}
	}
	return r.db.CreateEvent(ctx, dbgen.CreateEventParams{
		EventTypeKey:  params.EventTypeKey,
		DecisionID:    database.ToPgUUID(params.DecisionID),
		UserID:        params.UserID,
		Properties:    params.Properties,
		Timestamp:     params.Timestamp,
		ClientEventID: clientID,
	})
}

func (r *SQLCEventRepository) ExistsByTypeAndClientID(ctx context.Context, eventTypeKey, clientEventID string) (bool, error) {
	return r.db.ExistsEventByTypeAndClientID(ctx, dbgen.ExistsEventByTypeAndClientIDParams{
		EventTypeKey:  eventTypeKey,
		ClientEventID: database.ToPgText(&clientEventID),
	})
}

func (r *SQLCEventRepository) ListByDecisionIDsAndWindow(
	ctx context.Context,
	decisionIDs []uuid.UUID,
	from, to time.Time,
) ([]*models.Event, error) {
	fromPg := pgtype.Timestamptz{Time: from, Valid: true}
	toPg := pgtype.Timestamptz{Time: to, Valid: true}

	rows, err := r.db.ListEventsByDecisionIDsAndWindow(ctx, dbgen.ListEventsByDecisionIDsAndWindowParams{
		Column1:     decisionIDs,
		Timestamp:   fromPg,
		Timestamp_2: toPg,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*models.Event, len(rows))
	for i := range rows {
		out[i] = eventRowToDomain(rows[i])
	}

	return out, nil
}

func eventRowToDomain(row dbgen.Event) *models.Event {
	ev := &models.Event{
		ID:            row.ID,
		EventTypeKey:  row.EventTypeKey,
		UserID:        row.UserID,
		Properties:    row.Properties,
		DecisionID:    database.FromPgUUID(row.DecisionID),
		ClientEventID: database.FromPgText(row.ClientEventID),
	}
	if row.Timestamp.Valid {
		ev.Timestamp = row.Timestamp.Time
	}

	return ev
}
