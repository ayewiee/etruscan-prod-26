package logger

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func ZapLoggerMiddleware(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			latency := time.Since(start)

			fields := []zap.Field{
				zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
				zap.String("method", c.Request().Method),
				zap.String("uri", c.Request().RequestURI),
				zap.Int("status", c.Response().Status),
				zap.Duration("latency", latency),
				zap.String("remote_ip", c.RealIP()),
			}

			status := c.Response().Status
			if err == nil && status < 400 {
				log.Info("http_request", fields...)
			}

			return err
		}
	}
}
