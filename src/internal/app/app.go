package app

import (
	"context"
	"etruscan/internal/app/logger"
	"etruscan/internal/database"
	dbgen "etruscan/internal/database/generated"
	"etruscan/internal/infrastructure/cache"
	"etruscan/internal/provider"
	"etruscan/internal/repository"
	cacherepo "etruscan/internal/repository/cache"
	"etruscan/internal/usecases"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Context     context.Context
	Echo        *echo.Echo
	Config      *Config
	DB          *dbgen.Queries
	DBPool      *pgxpool.Pool
	RedisClient *cache.Client
	Deps        *Dependencies
}

func NewApiApp(ctx context.Context, cfg Config) (*App, error) {
	dbPool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	queries := dbgen.New(dbPool)
	redisClient := cache.NewRedisClient(cfg.RedisAddr)

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
	decisionRepo := repository.NewSQLCDecisionRepository(queries)
	eventRepo := repository.NewSQLCEventRepository(queries)
	eventTypeRepo := repository.NewSQLCEventTypeRepository(queries)
	metricRepo := repository.NewSQLCMetricRepository(queries)
	guardrailRepo := repository.NewSQLCGuardrailRepository(queries, metricRepo)

	runningExpCache := cacherepo.NewRunningExperimentCache(redisClient)
	ptcptnTracker := cacherepo.NewParticipationTracker(redisClient)
	flagCache := cacherepo.NewFlagCache(redisClient)

	metricComputer := usecases.NewMetricComputer(decisionRepo, eventRepo, eventTypeRepo)

	deps := Dependencies{
		AuthUseCase:          usecases.NewAuthUseCase(userRepo, passwordHasher, jwtProvider),
		UserUseCase:          usecases.NewUserUseCase(userRepo, passwordHasher),
		ApproverGroupUseCase: usecases.NewApproverGroupUseCase(approverGroupRepo, userRepo),
		FlagUseCase:          usecases.NewFlagUseCase(flagRepo),
		ExperimentUseCase: usecases.NewExperimentUseCase(
			experimentRepo,
			flagRepo,
			userRepo,
			approverGroupRepo,
			metricRepo,
			guardrailRepo,
			runningExpCache,
			cfg.DefaultMinApprovals,
		),
		DecideUseCase: usecases.NewDecideUseCase(
			runningExpCache,
			ptcptnTracker,
			flagCache,
			decisionRepo,
			experimentRepo,
			flagRepo,
			log,
		),
		EventsUseCase: usecases.NewEventsUseCase(eventRepo, eventTypeRepo),
		MetricUseCase: usecases.NewMetricUseCase(metricRepo),
		ReportUseCase: usecases.NewReportUseCase(
			experimentRepo,
			metricRepo,
			metricComputer,
		),
	}

	guardrailRunner := usecases.NewGuardrailRunner(
		experimentRepo,
		guardrailRepo,
		metricRepo,
		metricComputer,
		log,
		cfg.GuardrailCheckIntervalMinutes,
	)
	go guardrailRunner.Run(ctx)

	app := &App{
		Context:     ctx,
		Echo:        NewServer(log),
		Config:      &cfg,
		DB:          queries,
		DBPool:      dbPool,
		RedisClient: redisClient,
		Deps:        &deps,
	}

	app.RegisterRoutes()

	return app, nil
}
