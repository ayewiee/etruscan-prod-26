package usecases

import (
	"context"
	"etruscan/internal/domain/models"
	"etruscan/internal/repository"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationRouter is responsible for resolving which users should
// receive a notification and dispatching it to the appropriate channels.
type NotificationRouter struct {
	settingsRepo repository.NotificationSettingsRepository
	userRepo     repository.UserRepository

	telegramChannel models.NotificationChannel
	emailChannel    models.NotificationChannel

	logger *zap.Logger

	highInterval time.Duration
	lowInterval  time.Duration

	mu       sync.Mutex
	lastSent map[string]time.Time
}

func NewNotificationRouter(
	settingsRepo repository.NotificationSettingsRepository,
	userRepo repository.UserRepository,
	telegramChannel models.NotificationChannel,
	emailChannel models.NotificationChannel,
	logger *zap.Logger,
	highInterval time.Duration,
	lowInterval time.Duration,
) *NotificationRouter {
	return &NotificationRouter{
		settingsRepo:    settingsRepo,
		userRepo:        userRepo,
		telegramChannel: telegramChannel,
		emailChannel:    emailChannel,
		logger:          logger,
		highInterval:    highInterval,
		lowInterval:     lowInterval,
		lastSent:        make(map[string]time.Time),
	}
}

func (r *NotificationRouter) Notify(ctx context.Context, n models.Notification) {
	if r.telegramChannel == nil && r.emailChannel == nil {
		return
	}

	settings, err := r.settingsRepo.ListForExperiment(ctx, n.ExperimentID)
	if err != nil {
		r.logger.Error("notification router: list settings", zap.Error(err), zap.Any("experimentId", n.ExperimentID))
		return
	}

	if len(settings) == 0 {
		return
	}

	for _, setting := range settings {
		if !setting.Severity.ShouldBeDelivered(n.Severity) {
			continue
		}

		user, err := r.userRepo.GetById(ctx, setting.UserID)
		if err != nil {
			r.logger.Error("notification router: get user", zap.Error(err), zap.Any("userId", setting.UserID))
			continue
		}

		if r.isRateLimited(n, user.ID) {
			continue
		}

		// Telegram
		if setting.EnableTelegram && r.telegramChannel != nil && user.TelegramChatID != nil {
			if err := r.telegramChannel.Send(ctx, user, n); err != nil {
				r.logger.Error("notification router: telegram send failed",
					zap.Error(err),
					zap.Any("userId", user.ID),
					zap.Any("experimentId", n.ExperimentID),
				)
			}
		}

		// Email
		if setting.EnableEmail && r.emailChannel != nil {
			if err := r.emailChannel.Send(ctx, user, n); err != nil {
				r.logger.Error("notification router: email send failed",
					zap.Error(err),
					zap.Any("userId", user.ID),
					zap.Any("experimentId", n.ExperimentID),
				)
			}
		}
	}
}

func (r *NotificationRouter) isRateLimited(n models.Notification, userID uuid.UUID) bool {
	key := r.rateLimitKey(n, userID)

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	last, ok := r.lastSent[key]
	if !ok {
		r.lastSent[key] = now
		return false
	}

	var interval time.Duration
	if n.Severity == models.NotificationSeverityHigh {
		interval = r.highInterval
	} else {
		interval = r.lowInterval
	}

	if interval <= 0 {
		r.lastSent[key] = now
		return false
	}

	if now.Sub(last) < interval {
		// Rate-limited; do not update lastSent so the window slides from first send.
		r.logger.Debug("notification router: rate limited",
			zap.String("key", key),
			zap.Duration("interval", interval),
		)
		return true
	}

	r.lastSent[key] = now
	return false
}

func (r *NotificationRouter) rateLimitKey(n models.Notification, userID uuid.UUID) string {
	return string(n.Type) + "|" + n.ExperimentID.String() + "|" + userID.String()
}
