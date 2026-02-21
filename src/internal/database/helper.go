package database

import (
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgUUID(v *uuid.UUID) pgtype.UUID {
	if v == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *v, Valid: true}
}
func FromPgUUID(v pgtype.UUID) *uuid.UUID {
	if v.Valid {
		return (*uuid.UUID)(&v.Bytes)
	}
	return nil
}

func ToPgInt(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}
func FromPgInt(v pgtype.Int4) *int {
	if v.Valid {
		num := int(v.Int32)
		return &num
	}
	return nil
}

func ToPgText(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *v, Valid: true}
}
func FromPgText(v pgtype.Text) *string {
	if v.Valid {
		return &v.String
	}
	return nil
}

func FromPgTimestamptz(v pgtype.Timestamptz) *time.Time {
	if v.Valid {
		return &v.Time
	}
	return nil
}

func ToNullExperimentOutcome(v *models.ExperimentOutcome) dbgen.NullExperimentOutcome {
	if v == nil {
		return dbgen.NullExperimentOutcome{Valid: false}
	}
	return dbgen.NullExperimentOutcome{ExperimentOutcome: dbgen.ExperimentOutcome(*v), Valid: true}
}
func FromNullExperimentOutcome(v dbgen.NullExperimentOutcome) *models.ExperimentOutcome {
	if v.Valid {
		eo := models.ExperimentOutcome(v.ExperimentOutcome)
		return &eo
	}
	return nil
}

func ToNullExperimentStatus(v *models.ExperimentStatus) dbgen.NullExperimentStatus {
	if v == nil {
		return dbgen.NullExperimentStatus{Valid: false}
	}
	return dbgen.NullExperimentStatus{ExperimentStatus: dbgen.ExperimentStatus(*v), Valid: true}
}
func FromNullExperimentStatus(v dbgen.NullExperimentStatus) *models.ExperimentStatus {
	if v.Valid {
		es := models.ExperimentStatus(v.ExperimentStatus)
		return &es
	}
	return nil
}
