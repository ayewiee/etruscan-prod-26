package app

type Config struct {
	HttpPort            int    `env:"HTTP_PORT, default=8080"`
	DatabaseURL         string `env:"DATABASE_URL, required"`
	ProductionMode      bool   `env:"PRODUCTION_MODE, default=true"`
	JWTSecret           string `env:"JWT_SECRET, required"`
	DefaultMinApprovals int    `env:"DEFAULT_MIN_APPROVALS, default=1"`
}
