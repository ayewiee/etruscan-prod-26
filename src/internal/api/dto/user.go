package dto

import (
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type AdminCreateUpdateUserRequest struct {
	Email         string     `json:"email" validate:"required,email,max=100"`
	Password      *string    `json:"password" validate:"omitempty,required,min=8,max=64"`
	Username      string     `json:"username" validate:"required,min=4,max=32"`
	Role          string     `json:"role" validate:"required,oneof=VIEWER EXPERIMENTER APPROVER ADMIN"`
	MinApprovals  *int       `json:"minApprovals" validate:"omitempty,gte=1"`
	ApproverGroup *uuid.UUID `json:"approverGroup" validate:"omitempty,uuid"`
}

func (req *AdminCreateUpdateUserRequest) ToDomain() *models.User {
	return &models.User{
		ID:            uuid.UUID{},
		Email:         req.Email,
		Username:      req.Username,
		Role:          models.UserRole(req.Role),
		MinApprovals:  req.MinApprovals,
		ApproverGroup: req.ApproverGroup,
	}
}

type UserUpdateRequest struct {
	Username *string `json:"username" validate:"omitempty,required,min=4,max=32"`
	Password *string `json:"password" validate:"omitempty,required,min=8,max=64"`
}

type UserResponseDTO struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	Role          string     `json:"role"`
	MinApprovals  *int       `json:"minApprovals"`
	ApproverGroup *uuid.UUID `json:"approverGroup"`
	CreatedAt     string     `json:"createdAt"`
	UpdatedAt     string     `json:"updatedAt"`
}

type CompactUserDTO struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func UserResponseDTOFromDomain(u *models.User) UserResponseDTO {
	return UserResponseDTO{
		ID:            u.ID.String(),
		Email:         u.Email,
		Username:      u.Username,
		Role:          string(u.Role),
		MinApprovals:  u.MinApprovals,
		ApproverGroup: u.ApproverGroup,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     u.UpdatedAt.Format(time.RFC3339),
	}
}

func CompactUserDTOFromDomain(u *models.User) CompactUserDTO {
	return CompactUserDTO{
		ID:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		Role:     string(u.Role),
	}
}

func PaginatedUserResponseDTOListFromDomainList(ul []*models.User, total, page, size int) *PaginatedResponse {
	userDTOs := make([]UserResponseDTO, len(ul))
	for i, u := range ul {
		userDTOs[i] = UserResponseDTOFromDomain(u)
	}

	return &PaginatedResponse{
		Items: userDTOs,
		Total: total,
		Page:  page,
		Size:  size,
	}
}
