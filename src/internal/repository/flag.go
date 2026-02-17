package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type FlagRepository interface {
	Create(ctx context.Context, flag models.Flag) (models.Flag, error)
	Update(ctx context.Context, flag models.Flag) (models.Flag, error)
	GetByID(ctx context.Context, id uuid.UUID) (models.Flag, error)
	GetByKey(ctx context.Context, key string) (models.Flag, error)
	List(ctx context.Context) ([]models.Flag, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SQLCFlagRepository struct {
	db *dbgen.Queries
}

func NewSQLCFlagRepository(db *dbgen.Queries) FlagRepository {
	return &SQLCFlagRepository{db: db}
}

func (r *SQLCFlagRepository) Create(ctx context.Context, flag models.Flag) (models.Flag, error) {
	row, err := r.db.CreateFlag(ctx, dbgen.CreateFlagParams{
		Key:          flag.Key,
		Description:  database.ToPgText(flag.Description),
		DefaultValue: flag.DefaultValue,
		ValueType:    string(flag.ValueType),
	})
	if err != nil {
		return models.Flag{}, err
	}

	return flagRowToDomain(row), nil
}

func (r *SQLCFlagRepository) Update(ctx context.Context, flag models.Flag) (models.Flag, error) {
	row, err := r.db.UpdateFlag(ctx, dbgen.UpdateFlagParams{
		ID:           flag.ID,
		Key:          flag.Key,
		Description:  database.ToPgText(flag.Description),
		DefaultValue: flag.DefaultValue,
		ValueType:    string(flag.ValueType),
	})
	if err != nil {
		return models.Flag{}, err
	}

	return flagRowToDomain(row), nil
}

func (r *SQLCFlagRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Flag, error) {
	row, err := r.db.GetFlagByID(ctx, id)
	if err != nil {
		return models.Flag{}, err
	}
	return flagRowToDomain(row), nil
}

func (r *SQLCFlagRepository) GetByKey(ctx context.Context, key string) (models.Flag, error) {
	row, err := r.db.GetFlagByKey(ctx, key)
	if err != nil {
		return models.Flag{}, err
	}
	return flagRowToDomain(row), nil
}

func (r *SQLCFlagRepository) List(ctx context.Context) ([]models.Flag, error) {
	rows, err := r.db.ListFlags(ctx)
	if err != nil {
		return []models.Flag{}, err
	}

	flags := make([]models.Flag, len(rows))
	for i, row := range rows {
		flags[i] = flagRowToDomain(row)
	}
	return flags, nil
}

func (r *SQLCFlagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteFlag(ctx, id)
}

func flagRowToDomain(row dbgen.Flag) models.Flag {
	return models.Flag{
		ID:           row.ID,
		Key:          row.Key,
		Description:  database.FromPgText(row.Description),
		DefaultValue: row.DefaultValue,
		ValueType:    models.FlagValueType(row.ValueType),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
