package dto

import (
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
)

type CreateNotificationSettings struct {
	Severity       string `json:"severity" validate:"required,oneof=LOW HIGH"`
	EnableTelegram bool   `json:"enableTelegram"`
	EnableEmail    bool   `json:"enableEmail"`
}

type NotificationSettingsResponse struct {
	ID             uuid.UUID `json:"id"`
	ExperimentID   uuid.UUID `json:"experimentId"`
	Severity       string    `json:"severity"`
	EnableTelegram bool      `json:"enableTelegram"`
	EnableEmail    bool      `json:"enableEmail"`
}

func NotificationSettingsResponseFromDomain(ns *models.NotificationSettings) *NotificationSettingsResponse {
	return &NotificationSettingsResponse{
		ID:             ns.ID,
		ExperimentID:   ns.ExperimentID,
		Severity:       string(ns.Severity),
		EnableTelegram: ns.EnableTelegram,
		EnableEmail:    ns.EnableEmail,
	}
}

func NotificationSettingsResponseListFromDomain(settings []*models.NotificationSettings) []*NotificationSettingsResponse {
	res := make([]*NotificationSettingsResponse, len(settings))
	for i, setting := range settings {
		res[i] = NotificationSettingsResponseFromDomain(setting)
	}
	return res
}
