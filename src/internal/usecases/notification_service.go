package usecases

import (
	"context"
	"etruscan/internal/domain/models"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
)

// NotifyGuardrailTriggered builds and sends a HIGH severity notification
// when a guardrail triggers for an experiment.
func NotifyGuardrailTriggered(
	ctx context.Context,
	router *NotificationRouter,
	experiment *models.Experiment,
	guardrail *models.Guardrail,
	value float64,
) {
	if router == nil {
		return
	}

	title := fmt.Sprintf("⚠️Guardrail triggered for experiment _%s_", experiment.Name)
	body := fmt.Sprintf(
		"Guardrail on metric %s triggered for experiment _%s_ (```%s```).\n\nThreshold: %.4f\nValue: %.4f\nWindow (s): %d\nAction: %s",
		guardrail.MetricKey,
		experiment.Name,
		experiment.ID,
		guardrail.Threshold,
		value,
		guardrail.WindowSeconds,
		guardrail.Action,
	)

	n := models.Notification{
		Type:           models.NotificationEventTypeGuardrailTriggered,
		Severity:       models.NotificationSeverityHigh,
		ExperimentID:   experiment.ID,
		ExperimentName: experiment.Name,
		FlagKey:        experiment.FlagKey,
		Title:          title,
		Body:           body,
		Metadata: map[string]string{
			"metricKey":       guardrail.MetricKey,
			"experimentId":    experiment.ID.String(),
			"experimentName":  experiment.Name,
			"flagKey":         experiment.FlagKey,
			"threshold":       fmt.Sprintf("%f", guardrail.Threshold),
			"value":           fmt.Sprintf("%f", value),
			"thresholdWindow": fmt.Sprintf("%d", guardrail.WindowSeconds),
			"action":          guardrail.Action,
		},
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Fatalf("Recovered from panic in Notify: %v\n", r)
			}
		}()

		// goroutine has exactly 5 seconds to notify
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		router.Notify(ctxWithTimeout, n)
	}()
}

// NotifyExperimentStatusChangedUser is used when a human actor changes status.
func NotifyExperimentStatusChangedUser(
	ctx context.Context,
	router *NotificationRouter,
	experiment *models.Experiment,
	oldStatus *models.ExperimentStatus,
	newStatus models.ExperimentStatus,
	outcome *models.ExperimentOutcome,
) {
	notifyExperimentStatusChanged(ctx, router, experiment, oldStatus, newStatus, outcome, false)
}

// NotifyExperimentStatusChangedSystem is used when the system changes status (e.g. via guardrails).
func NotifyExperimentStatusChangedSystem(
	ctx context.Context,
	router *NotificationRouter,
	experiment *models.Experiment,
	oldStatus *models.ExperimentStatus,
	newStatus models.ExperimentStatus,
	outcome *models.ExperimentOutcome,
) {
	notifyExperimentStatusChanged(ctx, router, experiment, oldStatus, newStatus, outcome, true)
}

func notifyExperimentStatusChanged(
	ctx context.Context,
	router *NotificationRouter,
	experiment *models.Experiment,
	oldStatus *models.ExperimentStatus,
	newStatus models.ExperimentStatus,
	outcome *models.ExperimentOutcome,
	system bool,
) {
	if router == nil {
		return
	}

	severity := models.NotificationSeverityLow
	switch newStatus {
	case models.ExperimentStatusLaunched, models.ExperimentStatusPaused, models.ExperimentStatusFinished:
		severity = models.NotificationSeverityHigh
	}

	source := "User"
	if system {
		source = "System"
	}

	title := fmt.Sprintf("%s changed status of experiment %s to %s", source, experiment.Name, newStatus)

	fromStatusString := ""
	if oldStatus != nil {
		fromStatusString = fmt.Sprintf(" from %s", *oldStatus)
	}

	body := fmt.Sprintf(
		"Experiment %s (%s) status changed%s to %s.",
		experiment.Name,
		experiment.ID,
		fromStatusString,
		newStatus,
	)

	if outcome != nil {
		body += fmt.Sprintf("\nOutcome: %s.", *outcome)
	}

	metadata := map[string]string{
		"experimentId":   experiment.ID.String(),
		"experimentName": experiment.Name,
		"flagKey":        experiment.FlagKey,
		"newStatus":      string(newStatus),
		"source":         source,
	}
	if oldStatus != nil {
		metadata["oldStatus"] = string(*oldStatus)
	}
	if outcome != nil {
		metadata["outcome"] = string(*outcome)
	}

	n := models.Notification{
		Type:           models.NotificationEventTypeExperimentStatusChanged,
		Severity:       severity,
		ExperimentID:   experiment.ID,
		ExperimentName: experiment.Name,
		FlagKey:        experiment.FlagKey,
		Title:          title,
		Body:           body,
		Metadata:       metadata,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				router.logger.Error("Recovered from panic in Notify", zap.Error(fmt.Errorf("%v", r)))
			}
		}()

		// goroutine has exactly 5 seconds to notify
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		router.Notify(ctxWithTimeout, n)
	}()
}
