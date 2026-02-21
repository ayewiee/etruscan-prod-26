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

type FlagCache struct {
	rdb *cache.Client
}

func NewFlagCache(rdb *cache.Client) *FlagCache {
	return &FlagCache{rdb: rdb}
}

type cachedFlag struct {
	ID           uuid.UUID       `json:"id"`
	Key          string          `json:"key"`
	Description  *string         `json:"description,omitempty"`
	DefaultValue json.RawMessage `json:"defaultValue"`
	ValueType    string          `json:"valueType"`
}

const flagExpTTL = 10 * time.Minute

func (c *FlagCache) Get(ctx context.Context, flagKey string) (*models.Flag, error) {
	key := "flag:" + flagKey

	data, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}

	var cf cachedFlag
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, err
	}

	return &models.Flag{
		ID:           cf.ID,
		Key:          cf.Key,
		Description:  cf.Description,
		DefaultValue: cf.DefaultValue,
		ValueType:    models.FlagValueType(cf.ValueType),
	}, nil
}

func (c *FlagCache) Set(ctx context.Context, flag *models.Flag) error {
	key := "flag:" + flag.Key

	cf := cachedFlag{
		ID:           flag.ID,
		Key:          flag.Key,
		Description:  flag.Description,
		DefaultValue: flag.DefaultValue,
		ValueType:    string(flag.ValueType),
	}

	data, _ := json.Marshal(cf)
	return c.rdb.Set(ctx, key, data, flagExpTTL).Err()
}

func (c *FlagCache) Delete(ctx context.Context, flagKey string) error {
	return c.rdb.Del(ctx, "flag:"+flagKey).Err()
}
