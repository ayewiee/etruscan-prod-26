package repository

import (
	"context"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type ApproverGroupRepository interface {
	Create(ctx context.Context, approverGroup models.ApproverGroup) (models.ApproverGroup, error)
	GetByID(ctx context.Context, id uuid.UUID) (models.ApproverGroup, error)
	Update(ctx context.Context, approverGroup models.ApproverGroup) (models.ApproverGroup, error)
	AddMembers(ctx context.Context, id uuid.UUID, members []uuid.UUID) error
	RemoveMembers(ctx context.Context, id uuid.UUID, members []uuid.UUID) error
	List(ctx context.Context) ([]models.ApproverGroup, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SQLCApproverGroupRepository struct {
	db *dbgen.Queries
}

func NewSQLCApproverGroupRepository(db *dbgen.Queries) *SQLCApproverGroupRepository {
	return &SQLCApproverGroupRepository{db: db}
}

func (r SQLCApproverGroupRepository) Create(ctx context.Context, approverGroup models.ApproverGroup) (models.ApproverGroup, error) {
	row, err := r.db.CreateApproverGroup(ctx, dbgen.CreateApproverGroupParams{
		Name:        approverGroup.Name,
		Description: database.ToPgText(approverGroup.Description),
	})
	if err != nil {
		return models.ApproverGroup{}, err
	}

	return approverGroupRowToDomain(row), nil
}

func (r SQLCApproverGroupRepository) GetMembers(ctx context.Context, id uuid.UUID) ([]models.User, error) {
	rows, err := r.db.GetApproverGroupMembers(ctx, id)
	if err != nil {
		return []models.User{}, err
	}
	users := make([]models.User, len(rows))
	for i, row := range rows {
		users[i] = models.User{
			ID:       row.ID,
			Email:    row.Email,
			Username: row.Username,
			Role:     models.UserRole(row.Role),
		}
	}

	return users, nil
}

func (r SQLCApproverGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (models.ApproverGroup, error) {
	row, err := r.db.GetApproverGroup(ctx, id)
	if err != nil {
		return models.ApproverGroup{}, err
	}

	members, err := r.GetMembers(ctx, id)
	if err != nil {
		return models.ApproverGroup{}, err
	}

	ag := approverGroupRowToDomain(row)
	ag.Members = members

	return ag, nil
}

func (r SQLCApproverGroupRepository) Update(ctx context.Context, approverGroup models.ApproverGroup) (models.ApproverGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (r SQLCApproverGroupRepository) AddMembers(ctx context.Context, id uuid.UUID, members []uuid.UUID) error {
	return r.db.AddApproversToApproverGroup(ctx, dbgen.AddApproversToApproverGroupParams{
		Column1:         members,
		ApproverGroupID: id,
	})
}

func (r SQLCApproverGroupRepository) RemoveMembers(ctx context.Context, id uuid.UUID, members []uuid.UUID) error {
	return r.db.RemoveApproversFromApproverGroup(ctx, dbgen.RemoveApproversFromApproverGroupParams{
		Column2:         members,
		ApproverGroupID: id,
	})
}

func (r SQLCApproverGroupRepository) List(ctx context.Context) ([]models.ApproverGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (r SQLCApproverGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteApproverGroup(ctx, id)
}

func approverGroupRowToDomain(row dbgen.ApproverGroup) models.ApproverGroup {
	return models.ApproverGroup{
		ID:          row.ID,
		Name:        row.Name,
		Description: database.FromPgText(row.Description),
		Members:     []models.User{},
		CreatedAt:   row.CreatedAt.Time,
	}
}
