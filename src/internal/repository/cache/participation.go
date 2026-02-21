package cache

import (
	"context"
	"etruscan/internal/infrastructure/cache"
	"time"

	"github.com/google/uuid"
)

type ParticipationTracker struct {
	rdb *cache.Client
}

const maxActiveExperiments = 2
const participationTTL = 30 * 24 * time.Hour

func NewParticipationTracker(rdb *cache.Client) *ParticipationTracker {
	return &ParticipationTracker{rdb: rdb}
}

func (t *ParticipationTracker) CanParticipate(ctx context.Context, userID string, expID uuid.UUID) (bool, error) {
	key := "user_active_exps:" + userID

	count := t.rdb.SCard(ctx, key).Val()
	if count >= int64(maxActiveExperiments) {
		return false, nil
	}

	// Add this experiment to the set
	t.rdb.SAdd(ctx, key, expID.String())
	t.rdb.Expire(ctx, key, participationTTL)

	return true, nil
}
