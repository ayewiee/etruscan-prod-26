package usecases

import (
	"context"
	"database/sql"
	"errors"
	"etruscan/internal/domain/bucketing"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"etruscan/internal/repository/cache"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DecideUseCase struct {
	runningExpCache           *cache.RunningExperimentCache
	participationTrackerCache *cache.ParticipationTracker
	flagCache                 *cache.FlagCache
	decisionRepo              repository.DecisionRepository
	expRepo                   repository.ExperimentRepository
	flagRepo                  repository.FlagRepository
	logger                    *zap.Logger
}

func NewDecideUseCase(
	runningExpCache *cache.RunningExperimentCache,
	participationTrackerCache *cache.ParticipationTracker,
	flagCache *cache.FlagCache,
	decisionRepo repository.DecisionRepository,
	expRepo repository.ExperimentRepository,
	flagRepo repository.FlagRepository,
	logger *zap.Logger,
) *DecideUseCase {
	return &DecideUseCase{
		runningExpCache,
		participationTrackerCache,
		flagCache,
		decisionRepo,
		expRepo,
		flagRepo,
		logger,
	}
}

func (uc *DecideUseCase) getFlag(ctx context.Context, flagKey string) (*models.Flag, error) {
	flag, err := uc.flagCache.Get(ctx, flagKey)
	if err != nil {
		return nil, err
	}

	if flag == nil {
		flag, err = uc.flagRepo.GetByKey(ctx, flagKey)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		_ = uc.flagCache.Set(ctx, flag)
	}

	return flag, nil
}

func (uc *DecideUseCase) getRunningExperiment(ctx context.Context, flagKey string) (*models.Experiment, error) {
	exp, err := uc.runningExpCache.Get(ctx, flagKey)
	if err != nil {
		return nil, err
	}

	if exp == nil {
		exp, err = uc.expRepo.GetRunningExperimentByFlagKey(ctx, flagKey)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		variants, err := uc.expRepo.ListVariantsByExperimentID(ctx, exp.ID)
		if err != nil {
			return nil, err
		}
		exp.Variants = variants
		_ = uc.runningExpCache.Set(ctx, flagKey, exp)
	}

	return exp, nil
}

type DecideParams struct {
	UserID  string
	FlagKey string
	Context map[string]interface{}
}

func (uc *DecideUseCase) Decide(ctx context.Context, params DecideParams) (*models.Decision, error) {
	flag, err := uc.getFlag(ctx, params.FlagKey)
	if err != nil {
		return nil, err
	}

	// no safe-fail for a non-existent flag
	if flag == nil {
		return nil, models.NewErrNotFound(
			"Flag not found",
			map[string]interface{}{"key": params.FlagKey},
			nil,
		)
	}

	exp, err := uc.getRunningExperiment(ctx, params.FlagKey)
	if err != nil {
		return nil, err
	}

	// if there's no running experiment on this flag — return default value
	if exp == nil {
		uc.logger.Debug(
			"no running experiment — user got default value",
			zap.Any("value", flag.DefaultValue),
		)
		return uc.returnDefault(ctx, flag, params)
	}

	participationBucket, err := bucketing.HashAndBucket(params.UserID, flag.Key, exp.ID, nil)
	if err != nil { // fail-safe
		uc.logger.Error("failed to hash decision, returned defaultValue", zap.Error(err))
		uc.logger.Debug(
			"user got default value",
			zap.Any("participationBucket", nil),
			zap.Any("value", flag.DefaultValue),
		)
		return uc.returnDefault(ctx, flag, params)
	}

	// check audience percentage — e.g. bucket 62 >= 50% — client does not participate
	if participationBucket >= exp.AudiencePercentage {
		uc.logger.Debug(
			"user did not get into experiment audience — got default value",
			zap.Any("participationBucket", participationBucket),
			zap.Any("value", flag.DefaultValue),
		)
		return uc.returnDefault(ctx, flag, params)
	}

	canParticipate, err := uc.participationTrackerCache.CanParticipate(ctx, params.UserID, exp.ID)
	if err != nil {
		uc.logger.Error("participation check failed, returned default", zap.Error(err))
		return uc.returnDefault(ctx, flag, params)
	}
	if !canParticipate {
		uc.logger.Debug(
			"user participation limit reached — got default value",
			zap.String("userId", params.UserID),
			zap.Any("value", flag.DefaultValue),
		)
		return uc.returnDefault(ctx, flag, params)
	}

	variantSalt := "variant"
	variantBucket, err := bucketing.HashAndBucket(params.UserID, flag.Key, exp.ID, &variantSalt)
	if err != nil { // fail-safe
		uc.logger.Error("failed to hash for variant choice, returned defaultValue", zap.Error(err))
		uc.logger.Debug(
			"user got default value",
			zap.Any("participationBucket", nil),
			zap.Any("variantBucket", nil),
			zap.Any("value", flag.DefaultValue),
		)
		return uc.returnDefault(ctx, flag, params)
	}
	variant := bucketing.ChooseVariant(exp.Variants, variantBucket)

	uc.logger.Debug(
		"user participated in an experiment",
		zap.Any("participationBucket", participationBucket),
		zap.Any("variantBucket", variantBucket),
		zap.Any("value", variant.Value),
		zap.Any("variant", variant),
	)

	return uc.returnVariant(ctx, flag.Key, variant, params)
}

func (uc *DecideUseCase) returnVariant(
	ctx context.Context,
	flagKey string,
	variant *models.Variant,
	params DecideParams,
) (*models.Decision, error) {
	decision := &models.Decision{
		ExperimentID: &variant.ExperimentID,
		VariantID:    &variant.ID,
		FlagKey:      flagKey,
		Value:        variant.Value,
		UserID:       params.UserID,
		Context:      params.Context,
	}
	id, err := uc.decisionRepo.Create(ctx, decision)
	if err != nil {
		return nil, err
	}
	decision.ID = id

	return decision, nil
}

func (uc *DecideUseCase) returnDefault(ctx context.Context, flag *models.Flag, params DecideParams) (*models.Decision, error) {
	decision := &models.Decision{
		ExperimentID: nil,
		VariantID:    nil,
		FlagKey:      flag.Key,
		Value:        flag.DefaultValue,
		UserID:       params.UserID,
		Context:      params.Context,
	}
	id, err := uc.decisionRepo.Create(ctx, decision)
	if err != nil {
		return nil, err
	}
	decision.ID = id

	return decision, nil
}

func (uc *DecideUseCase) saveDecision(ctx context.Context, decision *models.Decision) (uuid.UUID, error) {
	return uc.decisionRepo.Create(ctx, decision)
}
