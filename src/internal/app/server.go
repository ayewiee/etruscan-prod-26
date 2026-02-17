package app

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/app/logger"
	"etruscan/internal/infrastructure/validator"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func NewServer(log *zap.Logger) *echo.Echo {
	e := echo.New()

	e.HTTPErrorHandler = apierrors.GlobalErrorHandler(log)
	e.Validator = validator.NewValidator()

	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.New().String()
		},
	}))

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		Skipper:           middleware.DefaultSkipper,
		StackSize:         4 << 10, // 4 KB
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			log.Error("panic recovered",
				zap.Error(err),
				zap.ByteString("stack", stack),
				zap.String("uri", c.Request().RequestURI),
			)
			return err
		},
	}))

	e.Use(logger.ZapLoggerMiddleware(log))

	return e
}
