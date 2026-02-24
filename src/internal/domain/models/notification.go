package models

import (
	"context"

	"github.com/google/uuid"
)

type NotificationSeverity string

const (
	NotificationSeverityLow  NotificationSeverity = "LOW"
	NotificationSeverityHigh NotificationSeverity = "HIGH"
)

func (ns NotificationSeverity) ShouldBeDelivered(eventSeverity NotificationSeverity) bool {
	switch ns {
	case NotificationSeverityLow:
		// LOW means receive all events, both low and high.
		return true
	case NotificationSeverityHigh:
		// HIGH means receive only high-severity events.
		return eventSeverity == NotificationSeverityHigh
	default:
		// Fallback to conservative behavior: treat as HIGH-only.
		return eventSeverity == NotificationSeverityHigh
	}
}

type NotificationEventType string

const (
	NotificationEventTypeGuardrailTriggered      NotificationEventType = "GUARDRAIL_TRIGGERED"
	NotificationEventTypeExperimentStatusChanged NotificationEventType = "EXPERIMENT_STATUS_CHANGED"
)

// Notification is a normalized message that can be sent to one or more channels.
type Notification struct {
	Type           NotificationEventType
	Severity       NotificationSeverity
	ExperimentID   uuid.UUID
	ExperimentName string
	FlagKey        string

	Title    string
	Body     string
	Metadata map[string]string
}

// NotificationChannel is implemented by concrete delivery channels (Telegram, Email, etc).
// Channels should not perform any DB lookups; they operate only on the provided user and notification.
type NotificationChannel interface {
	Send(ctx context.Context, user *User, n Notification) error
}
