package models

import (
	"time"

	"github.com/google/uuid"
)

type ApproverGroup struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Members     []*User
	CreatedAt   time.Time
}
