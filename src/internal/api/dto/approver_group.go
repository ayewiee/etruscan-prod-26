package dto

import (
	"etruscan/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type CreateApproverGroupRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
}

type ModifyApproverGroupMembersRequest struct {
	Users []uuid.UUID `json:"users" validate:"required,min=1,dive,required"`
}

type ApproverGroupResponse struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	Members     []CompactUserDTO `json:"members"`
	CreatedAt   string           `json:"createdAt"`
}

func ApproverGroupResponseFromDomain(ag *models.ApproverGroup) *ApproverGroupResponse {
	memberDTOs := make([]CompactUserDTO, len(ag.Members))
	for i, m := range ag.Members {
		memberDTOs[i] = CompactUserDTOFromDomain(&m)
	}

	return &ApproverGroupResponse{
		ID:          ag.ID,
		Name:        ag.Name,
		Description: ag.Description,
		Members:     memberDTOs,
		CreatedAt:   ag.CreatedAt.Format(time.RFC3339),
	}
}
