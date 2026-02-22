package usecases

import (
	"context"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GuardrailRunner periodically evaluates guardrails for LAUNCHED experiments and triggers pause/rollback.
type GuardrailRunner struct {
	experimentRepo       repository.ExperimentRepository
	guardrailRepo        repository.GuardrailRepository
	metricRepo           repository.MetricRepository
	computer             *MetricComputer
	logger               *zap.Logger
	checkIntervalMinutes int
}

func NewGuardrailRunner(
	experimentRepo repository.ExperimentRepository,
	guardrailRepo repository.GuardrailRepository,
	metricRepo repository.MetricRepository,
	computer *MetricComputer,
	logger *zap.Logger,
	checkIntervalMinutes int,
) *GuardrailRunner {
	return &GuardrailRunner{
		experimentRepo:       experimentRepo,
		guardrailRepo:        guardrailRepo,
		metricRepo:           metricRepo,
		computer:             computer,
		logger:               logger,
		checkIntervalMinutes: checkIntervalMinutes,
	}
}

func (r *GuardrailRunner) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(r.checkIntervalMinutes) * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.evaluateAll(ctx)
		}
	}
}

func (r *GuardrailRunner) evaluateAll(ctx context.Context) {
	experiments, _, err := r.experimentRepo.List(ctx, repository.ExperimentListFilters{
		Status: ptr(models.ExperimentStatusLaunched),
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		r.logger.Error("guardrail runner: list experiments", zap.Error(err))
		return
	}

	for _, exp := range experiments {
		r.evaluateExperiment(ctx, exp.ID)
	}
	r.logger.Info("guardrails check ran successfully", zap.Int("count", len(experiments)))
}

func (r *GuardrailRunner) evaluateExperiment(ctx context.Context, experimentID uuid.UUID) {
	guardrails, err := r.guardrailRepo.ListByExperimentID(ctx, experimentID)
	if err != nil {
		r.logger.Error(
			"guardrail runner: list guardrails",
			zap.String("experimentId", experimentID.String()),
			zap.Error(err),
		)
		return
	}

	r.logger.Debug(
		"\nEVALUATING EXPERIMENT\n",
		zap.Any("guardrails", guardrails),
		zap.Any("experimentID", experimentID),
	)

	for _, g := range guardrails {
		metric, err := r.metricRepo.GetByID(ctx, g.MetricID)
		if err != nil {
			continue
		}

		r.logger.Debug("EVALUATING METRIC", zap.Any("metric", metric))

		now := time.Now()
		from := now.Add(-time.Duration(g.WindowSeconds) * time.Second)

		val, err := r.computer.Compute(ctx, experimentID, nil, metric, from, now)
		if err != nil {
			continue
		}

		r.logger.Debug("EVALUATED METRIC VALUE", zap.Any("value", val), zap.Any("threshold", g.Threshold))

		triggered := false
		if g.ThresholdDirection == "upper" && val > g.Threshold {
			triggered = true
		}
		if g.ThresholdDirection == "lower" && val < g.Threshold {
			triggered = true
		}
		if !triggered {
			continue
		}

		switch g.Action {
		case "pause":
			_ = r.experimentRepo.UpdateStatus(ctx, experimentID, models.ExperimentStatusPaused)
		case "rollback":
			_, _ = r.experimentRepo.Finish(
				ctx,
				experimentID,
				models.ExperimentOutcomeRollback,
				"Guardrail with id "+g.ID.String()+" triggered",
				nil, // system -> nil
			)
		}

		_, err = r.guardrailRepo.CreateTrigger(ctx, g.ID, experimentID, val, g.MetricKey, g.Threshold, g.WindowSeconds, g.Action)
		if err != nil {
			r.logger.Error("guardrail runner: create trigger", zap.Error(err))
			return
		}

		r.logger.Info("guardrail triggered", zap.String("experimentId", experimentID.String()), zap.String("metricKey", g.MetricKey), zap.Float64("value", val), zap.Float64("threshold", g.Threshold))
	}
}

func ptr(s models.ExperimentStatus) *models.ExperimentStatus { return &s }
