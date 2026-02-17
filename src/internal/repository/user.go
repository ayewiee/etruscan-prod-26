package repository

import (
	"context"
	"errors"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) (models.User, error)
	GetById(ctx context.Context, id uuid.UUID) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	List(ctx context.Context, limit, offset int) ([]models.User, int, error)
	ValidateApproversExistenceAndRole(ctx context.Context, ids []uuid.UUID) (bool, error)
	Update(ctx context.Context, user models.User) (models.User, error)
	AdminUpdate(ctx context.Context, user models.User) (models.User, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

func NewSQLCUserRepository(db *dbgen.Queries) UserRepository {
	return SQLCUserRepository{db: db}
}

type SQLCUserRepository struct {
	db *dbgen.Queries
}

func (r SQLCUserRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	row, err := r.db.CreateUser(ctx, dbgen.CreateUserParams{
		Email:         user.Email,
		Username:      user.Username,
		PasswordHash:  user.PasswordHash,
		Role:          dbgen.UserRole(user.Role),
		MinApprovals:  database.ToPgInt(user.MinApprovals),
		ApproverGroup: database.ToPgUUID(user.ApproverGroup),
	})
	if err != nil {
		return models.User{}, err
	}

	return userRowToDomain(row), nil
}

func (r SQLCUserRepository) GetById(ctx context.Context, id uuid.UUID) (models.User, error) {
	row, err := r.db.GetUserById(ctx, id)
	if err != nil {
		return models.User{}, err
	}

	return userRowToDomain(row), nil
}

func (r SQLCUserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	row, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return models.User{}, err
	}

	return userRowToDomain(row), nil
}

func (r SQLCUserRepository) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	total, err := r.db.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []models.User{}, 0, nil
	}

	rows, err := r.db.ListUsers(ctx, dbgen.ListUsersParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return []models.User{}, 0, err
	}

	users := make([]models.User, len(rows))
	for i, row := range rows {
		users[i] = userRowToDomain(row)
	}

	return users, int(total), nil
}

var SomeUsersDoNotExistErr = errors.New("some users do not exist")

func (r SQLCUserRepository) ValidateApproversExistenceAndRole(ctx context.Context, ids []uuid.UUID) (bool, error) {
	return r.db.ValidateApproversExistAndRole(ctx, ids)
}

func (r SQLCUserRepository) Update(ctx context.Context, user models.User) (models.User, error) {
	row, err := r.db.UpdateUser(ctx, dbgen.UpdateUserParams{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return models.User{}, err
	}

	return userRowToDomain(row), nil
}

func (r SQLCUserRepository) AdminUpdate(ctx context.Context, user models.User) (models.User, error) {
	row, err := r.db.AdminUpdateUser(ctx, dbgen.AdminUpdateUserParams{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		PasswordHash:  user.PasswordHash,
		Role:          dbgen.UserRole(user.Role),
		MinApprovals:  database.ToPgInt(user.MinApprovals),
		ApproverGroup: database.ToPgUUID(user.ApproverGroup),
	})
	if err != nil {
		return models.User{}, err
	}

	return userRowToDomain(row), nil
}

func (r SQLCUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.SoftDeleteUser(ctx, id)
}

func userRowToDomain(row dbgen.User) models.User {
	return models.User{
		ID:            row.ID,
		Email:         row.Email,
		Username:      row.Username,
		PasswordHash:  row.PasswordHash,
		Role:          models.UserRole(row.Role),
		MinApprovals:  database.FromPgInt(row.MinApprovals),
		ApproverGroup: database.FromPgUUID(row.ApproverGroup),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}
