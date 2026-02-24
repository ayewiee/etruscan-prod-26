package models

import "github.com/google/uuid"

type NotificationSettings struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	ExperimentID   uuid.UUID
	Severity       NotificationSeverity // low/high
	EnableTelegram bool
	EnableEmail    bool
}
