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

func (r UserRole) CanEditFlags() bool {
	return r == UserRoleExperimenter || r == UserRoleAdmin
}

func (r UserRole) CanEditExperiments() bool {
	return r == UserRoleExperimenter || r == UserRoleAdmin
}

func (r UserRole) CanApproveInGeneral() bool {
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
