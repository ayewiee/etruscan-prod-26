package models

import (
	"time"

	"github.com/google/uuid"
)

type EventType struct {
	ID          uuid.UUID
	Key         string
	Name        string
	Description *string
	RequiresID  *uuid.UUID // FK to another event_type (e.g. exposure) for attribution
	RequiresKey *string
	CreatedAt   time.Time
}

type Event struct {
	ID            uuid.UUID
	EventTypeKey  string
	DecisionID    *uuid.UUID
	UserID        string
	Properties    []byte
	Timestamp     time.Time
	ClientEventID *string
}

type BatchEventItem struct {
	EventID      string
	EventTypeKey string
	DecisionID   uuid.UUID
	UserID       string
	Timestamp    time.Time
	Properties   map[string]interface{}
}

type BatchEventsResult struct {
	Accepted   int
	Duplicates int
	Rejected   int
	Errors     []BatchEventErr
}

type BatchEventErr struct {
	Index   int
	Message string
}
