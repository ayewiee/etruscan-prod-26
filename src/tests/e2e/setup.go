package e2e

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"etruscan/internal/app"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	BasePath     = "/api/v1"
	TestPort     = 18080
	AdminEmail   = "admin@etruscan.com"
	AdminPass    = "admin"
	E2EJWTSecret = "e2e-jwt-secret"
)

func StartPostgres(ctx context.Context, t *testing.T) (testcontainers.Container, string, error) {
	pgC, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("etruscan"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, "", err
	}
	host, _ := pgC.Host(ctx)
	port, _ := pgC.MappedPort(ctx, "5432")
	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:%s/etruscan?sslmode=disable", host, port.Port())
	return pgC, connStr, nil
}

func StartRedis(ctx context.Context, t *testing.T) (testcontainers.Container, string, error) {
	redisC, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections").WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		return nil, "", err
	}
	host, _ := redisC.Host(ctx)
	port, _ := redisC.MappedPort(ctx, "6379")
	return redisC, fmt.Sprintf("%s:%s", host, port.Port()), nil
}

func RunGoose(command, dbString, migrationsDir string) {
	absDir := migrationsDir
	if !filepath.IsAbs(migrationsDir) {
		cwd, _ := os.Getwd()
		for _, base := range []string{cwd, filepath.Join(cwd, "src")} {
			p := filepath.Join(base, migrationsDir)
			if _, err := os.Stat(p); err == nil {
				absDir = p
				break
			}
		}
	}
	cmd := exec.Command("goose", "postgres", dbString, command)
	cmd.Dir = absDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Goose %s failed:\n%s\nError: %v", command, out, err)
	}
}

func WaitForReady(t *testing.T, baseURL string) {
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 60; i++ {
		resp, err := client.Get(baseURL + "/ready")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout waiting for /ready")
}

func StartAPI(t *testing.T, ctx context.Context, connStr, redisAddr string) *app.App {
	cfg := app.Config{
		HttpPort:                      TestPort,
		DatabaseURL:                   connStr,
		RedisAddr:                     redisAddr,
		ProductionMode:                false,
		JWTSecret:                     E2EJWTSecret,
		DefaultMinApprovals:           1,
		GuardrailCheckIntervalMinutes: 2,
	}
	appInstance, err := app.NewApiApp(ctx, cfg)
	if err != nil {
		t.Fatalf("NewApiApp: %v", err)
	}
	return appInstance
}
