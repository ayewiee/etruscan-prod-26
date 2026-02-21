package cache

import (
	"context"
	"encoding/json"
	"errors"
	"etruscan/internal/domain/models"
	"etruscan/internal/infrastructure/cache"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RunningExperimentCache struct {
	rdb *cache.Client
}

func NewRunningExperimentCache(rdb *cache.Client) *RunningExperimentCache {
	return &RunningExperimentCache{rdb: rdb}
}

type cachedExperiment struct {
	ID                 uuid.UUID         `json:"id"`
	Version            int               `json:"version"`
	AudiencePercentage int               `json:"audience_pct"`
	TargetingRule      *string           `json:"targeting_rule,omitempty"`
	Variants           []*models.Variant `json:"variants"`
}

const activeExpTTL = 10 * time.Minute

func (c *RunningExperimentCache) Get(ctx context.Context, flagKey string) (*models.Experiment, error) {
	key := "active_exp:" + flagKey

	data, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}

	var ce cachedExperiment
	if err := json.Unmarshal(data, &ce); err != nil {
		return nil, err
	}

	return &models.Experiment{
		ID:                 ce.ID,
		Version:            ce.Version,
		AudiencePercentage: ce.AudiencePercentage,
		TargetingRule:      ce.TargetingRule,
		Variants:           ce.Variants,
	}, nil
}

func (c *RunningExperimentCache) Set(ctx context.Context, flagKey string, exp *models.Experiment) error {
	key := "active_exp:" + flagKey

	ce := cachedExperiment{
		ID:                 exp.ID,
		Version:            exp.Version,
		AudiencePercentage: exp.AudiencePercentage,
		TargetingRule:      exp.TargetingRule,
		Variants:           exp.Variants,
	}

	data, _ := json.Marshal(ce)
	return c.rdb.Set(ctx, key, data, activeExpTTL).Err()
}

func (c *RunningExperimentCache) Delete(ctx context.Context, flagKey string) error {
	return c.rdb.Del(ctx, "active_exp:"+flagKey).Err()
}
