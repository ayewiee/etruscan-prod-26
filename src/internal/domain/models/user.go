package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleAdmin        UserRole = "ADMIN"
	UserRoleApprover     UserRole = "APPROVER"
	UserRoleExperimenter UserRole = "EXPERIMENTER"
	UserRoleViewer       UserRole = "VIEWER"
)

func (r UserRole) CanManageFlags() bool {
	return r == UserRoleExperimenter || r == UserRoleAdmin
}

func (r UserRole) CanManageEventTypes() bool {
	return r == UserRoleExperimenter || r == UserRoleAdmin
}

func (r UserRole) CanManageExperiments() bool {
	return r == UserRoleExperimenter || r == UserRoleAdmin
}

func (r UserRole) CanApprove() bool {
	return r == UserRoleApprover || r == UserRoleAdmin
}

type User struct {
	ID            uuid.UUID
	Email         string
	Username      string
	PasswordHash  string
	Role          UserRole
	MinApprovals  *int
	ApproverGroup *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserAuthData struct {
	ID   uuid.UUID
	Role UserRole
}
