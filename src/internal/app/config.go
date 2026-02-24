package app

type Config struct {
	HttpPort                      int    `env:"HTTP_PORT, default=8080"`
	DatabaseURL                   string `env:"DATABASE_URL, required"`
	RedisAddr                     string `env:"REDIS_ADDR, required"`
	ProductionMode                bool   `env:"PRODUCTION_MODE, default=true"`
	JWTSecret                     string `env:"JWT_SECRET, required"`
	DefaultMinApprovals           int    `env:"DEFAULT_MIN_APPROVALS, default=1"`
	GuardrailCheckIntervalMinutes int    `env:"GUARDRAIL_CHECK_INTERVAL_MINUTES, default=2"`

	NotificationsEnabled   bool   `env:"NOTIFICATIONS_ENABLED, default=false"`
	NotifyTelegramEnabled  bool   `env:"NOTIFY_TELEGRAM_ENABLED, default=false"`
	NotifyTelegramBotToken string `env:"NOTIFY_TELEGRAM_BOT_TOKEN, required"`
	NotifyEmailEnabled     bool   `env:"NOTIFY_EMAIL_ENABLED, default=false"`
	NotifyEmailFrom        string `env:"NOTIFY_EMAIL_FROM, default=noreply@etruscan.com"`

	NotifyMinIntervalSecondsHigh int `env:"NOTIFY_MIN_INTERVAL_SECONDS_HIGH, default=60"`
	NotifyMinIntervalSecondsLow  int `env:"NOTIFY_MIN_INTERVAL_SECONDS_LOW, default=300"`
}
