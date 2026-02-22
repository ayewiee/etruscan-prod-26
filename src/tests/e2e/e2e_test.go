package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e in short mode")
	}

	ctx := context.Background()

	pgC, connStr, err := StartPostgres(ctx, t)
	require.NoError(t, err)
	defer func() { _ = pgC.Terminate(ctx) }()

	redisC, redisAddr, err := StartRedis(ctx, t)
	require.NoError(t, err)
	defer func() { _ = redisC.Terminate(ctx) }()

	migrationsDir := os.Getenv("E2E_MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "../../db/migrations"
	}
	RunGoose("up", connStr, migrationsDir)

	appInstance := StartAPI(t, ctx, connStr, redisAddr)
	defer appInstance.DBPool.Close()
	defer appInstance.RedisClient.Close()

	go func() {
		_ = appInstance.Echo.Start(fmt.Sprintf(":%d", TestPort))
	}()
	defer func() { _ = appInstance.Echo.Shutdown(context.Background()) }()

	baseURL := fmt.Sprintf("http://localhost:%d%s", TestPort, BasePath)
	WaitForReady(t, baseURL)

	client := &http.Client{Timeout: 10 * time.Second}

	// Login once to get token for authenticated tests
	resp, err := PostJSON(client, baseURL+"/auth/login", map[string]string{"email": AdminEmail, "password": AdminPass})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "login to get token")
	var loginOut struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginOut))
	require.NotEmpty(t, loginOut.Token, "token")
	token := loginOut.Token

	RunHealthAuthTests(t, client, baseURL, token)
	RunFlagsDecideTests(t, client, baseURL, token)
	RunExperimentLifecycleTests(t, client, baseURL, token)
	RunEventsTests(t, client, baseURL, token)
	RunReportGuardrailTests(t, client, baseURL, token)
}
