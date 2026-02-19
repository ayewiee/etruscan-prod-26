package app

import (
	"context"
	"etruscan/internal/app/logger"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/provider"
	"etruscan/internal/repository"
	"etruscan/internal/usecases"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Context context.Context
	Echo    *echo.Echo
	Config  *Config
	DB      *dbgen.Queries
	DBPool  *pgxpool.Pool
	Deps    *Dependencies
}

func NewApiApp(ctx context.Context, cfg Config) (*App, error) {
	dbPool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	queries := dbgen.New(dbPool)

	log, err := logger.NewZapLogger(cfg.ProductionMode)
	if err != nil {
		return nil, err
	}

	passwordHasher := provider.NewBcryptHasher(bcrypt.DefaultCost)
	jwtProvider := provider.NewJWTProvider([]byte(cfg.JWTSecret), 60*time.Minute)

	userRepo := repository.NewSQLCUserRepository(queries)
	approverGroupRepo := repository.NewSQLCApproverGroupRepository(queries)
	flagRepo := repository.NewSQLCFlagRepository(queries)
	experimentRepo := repository.NewSQLCExperimentRepository(queries)

	deps := Dependencies{
		AuthUseCase:          usecases.NewAuthUseCase(userRepo, passwordHasher, jwtProvider),
		UserUseCase:          usecases.NewUserUseCase(userRepo, passwordHasher),
		ApproverGroupUseCase: usecases.NewApproverGroupUseCase(approverGroupRepo, userRepo),
		FlagUseCase:          usecases.NewFlagUseCase(flagRepo),
		ExperimentUseCase:    usecases.NewExperimentUseCase(experimentRepo, flagRepo, userRepo, cfg.DefaultMinApprovals),
	}

	app := &App{
		Context: ctx,
		Echo:    NewServer(log),
		Config:  &cfg,
		DB:      queries,
		DBPool:  dbPool,
		Deps:    &deps,
	}

	app.RegisterRoutes()

	return app, nil
}
