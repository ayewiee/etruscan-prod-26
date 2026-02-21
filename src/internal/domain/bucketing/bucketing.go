package bucketing

import (
	"errors"
	"etruscan/internal/domain/models"

	"github.com/google/uuid"
	"github.com/spaolacci/murmur3"
)

// HashAndBucket returns a deterministic integer in range [0, 99].
// With same userID, flagKey, experimentID, user always falls into the same bucket.
// This guarantees stickiness.
func HashAndBucket(userID, flagKey string, experimentID uuid.UUID, additionalSalt *string) (int, error) {
	// Combine with separators that cannot appear in normal IDs
	key := userID + "|" + flagKey + "|" + experimentID.String()
	if additionalSalt != nil {
		key += "|" + *additionalSalt
	}

	hasher := murmur3.New64()
	_, err := hasher.Write([]byte(key))
	if err != nil {
		return 0, errors.New("hash write failed")
	}
	hash := hasher.Sum64()

	// 100 buckets (0-99) = 1% precision
	return int(hash % 100), nil
}

// ChooseVariant returns the variant for a given bucket (0-99).
// Uses cumulative weights → deterministic.
func ChooseVariant(variants []*models.Variant, bucket int) *models.Variant {
	cumulative := 0
	for i := range variants {
		cumulative += variants[i].Weight
		if bucket < cumulative {
			return variants[i]
		}
	}
	// fallback (should never happen if weights sum to 100)
	return variants[0]
}
