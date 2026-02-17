package database

import (
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
