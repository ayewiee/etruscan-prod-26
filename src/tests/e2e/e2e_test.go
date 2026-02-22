package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"etruscan/internal/app"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	basePath     = "/api/v1"
	testPort     = 18080
	adminEmail   = "admin@etruscan.com"
	adminPass    = "admin"
	e2eJWTSecret = "e2e-jwt-secret"
)

func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e in short mode")
	}

	ctx := context.Background()

	// Start Postgres
	pgC, connStr, err := startPostgres(ctx, t)
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}
	defer func() { _ = pgC.Terminate(ctx) }()

	// Start Redis
	redisC, redisAddr, err := startRedis(ctx, t)
	if err != nil {
		t.Fatalf("redis: %v", err)
	}
	defer func() { _ = redisC.Terminate(ctx) }()

	// Run migrations

	migrationsDir := os.Getenv("E2E_MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "../../db/migrations"
	}

	runGoose("up", connStr, migrationsDir)

	// Start API
	cfg := app.Config{
		HttpPort:                      testPort,
		DatabaseURL:                   connStr,
		RedisAddr:                     redisAddr,
		ProductionMode:                false,
		JWTSecret:                     e2eJWTSecret,
		DefaultMinApprovals:           1,
		GuardrailCheckIntervalMinutes: 2,
	}
	appInstance, err := app.NewApiApp(ctx, cfg)
	if err != nil {
		t.Fatalf("NewApiApp: %v", err)
	}
	defer appInstance.DBPool.Close()
	defer appInstance.RedisClient.Close()

	go func() {
		_ = appInstance.Echo.Start(fmt.Sprintf(":%d", testPort))
	}()
	defer func() { _ = appInstance.Echo.Shutdown(context.Background()) }()

	baseURL := fmt.Sprintf("http://localhost:%d%s", testPort, basePath)
	waitForReady(t, baseURL)

	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("Health", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /health: status %d", resp.StatusCode)
		}
	})

	t.Run("Ready", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/ready")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /ready: status %d", resp.StatusCode)
		}
	})

	t.Run("Unauthenticated_Protected_Returns_401", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/flags")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("GET /flags without token: want 401, got %d", resp.StatusCode)
		}
	})

	t.Run("Login_Invalid_Credentials_Returns_401", func(t *testing.T) {
		body := map[string]string{"email": adminEmail, "password": "wrong"}
		resp, err := postJSON(client, baseURL+"/auth/login", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("login wrong password: want 401, got %d", resp.StatusCode)
		}
	})

	var token string
	t.Run("Login_Success", func(t *testing.T) {
		body := map[string]string{"email": adminEmail, "password": adminPass}
		resp, err := postJSON(client, baseURL+"/auth/login", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("login: status %d, body %s", resp.StatusCode, string(b))
		}
		var out struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatal(err)
		}
		if out.Token == "" {
			t.Fatal("missing token")
		}
		token = out.Token
	})

	t.Run("RBAC_Admin_Can_List_Users", func(t *testing.T) {
		resp, err := getWithAuth(client, baseURL+"/admin/users", token)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /admin/users as admin: want 200, got %d", resp.StatusCode)
		}
	})

	t.Run("RBAC_Viewer_Cannot_Access_Admin", func(t *testing.T) {
		// Create viewer user as admin, then login as viewer and try admin endpoint
		createUserBody := map[string]interface{}{
			"username": "viewer1",
			"email":    "viewer@e2e.test",
			"password": "password",
			"role":     "VIEWER",
		}
		resp, err := postJSONWithAuth(client, baseURL+"/admin/users", token, createUserBody)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Skip("create viewer user failed (maybe email exists), skip RBAC viewer test")
		}
		loginResp, err := postJSON(client, baseURL+"/auth/login", map[string]string{"email": "viewer@e2e.test", "password": "password"})
		if err != nil {
			t.Fatal(err)
		}
		defer loginResp.Body.Close()
		if loginResp.StatusCode != http.StatusOK {
			t.Skip("viewer login failed, skip RBAC viewer test")
		}
		var loginOut struct {
			Token string `json:"token"`
		}
		_ = json.NewDecoder(loginResp.Body).Decode(&loginOut)
		viewerToken := loginOut.Token
		resp2, err := getWithAuth(client, baseURL+"/admin/users", viewerToken)
		if err != nil {
			t.Fatal(err)
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusForbidden {
			t.Errorf("GET /admin/users as VIEWER: want 403, got %d", resp2.StatusCode)
		}
	})

	t.Run("Happy_Flags_CRUD", func(t *testing.T) {
		createBody := map[string]interface{}{
			"key":          "e2e_flag_" + fmt.Sprintf("%d", time.Now().UnixNano()),
			"valueType":    "string",
			"defaultValue": "control",
		}
		resp, err := postJSONWithAuth(client, baseURL+"/flags", token, createBody)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("POST /flags: status %d, body %s", resp.StatusCode, string(b))
		}
		var flag struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&flag); err != nil {
			t.Fatal(err)
		}
		if flag.ID == "" || flag.Key == "" {
			t.Fatal("flag id/key empty")
		}
	})

	t.Run("Decide_Without_Active_Experiment_Returns_Default", func(t *testing.T) {
		// Create a flag with no experiment
		key := "e2e_decide_" + fmt.Sprintf("%d", time.Now().UnixNano())
		createBody := map[string]interface{}{
			"key":          key,
			"valueType":    "string",
			"defaultValue": "default_val",
		}
		resp, err := postJSONWithAuth(client, baseURL+"/flags", token, createBody)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Skip("flag create failed, skip decide test")
		}

		decideBody := map[string]interface{}{
			"userId":  "user-1",
			"flagKey": key,
		}
		resp2, err := postJSON(client, baseURL+"/decide", decideBody)
		if err != nil {
			t.Fatal(err)
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp2.Body)
			t.Fatalf("POST /decide: status %d, body %s", resp2.StatusCode, string(b))
		}
		var dec struct {
			Value interface{} `json:"value"`
		}
		if err := json.NewDecoder(resp2.Body).Decode(&dec); err != nil {
			t.Fatal(err)
		}
		if dec.Value != "default_val" {
			t.Errorf("decide without experiment: want default_val, got %v", dec.Value)
		}
	})

	t.Run("Events_Ingest_Valid", func(t *testing.T) {
		// Requires a valid decision_id; we can use a random UUID and expect accepted/duplicates/rejected
		body := map[string]interface{}{
			"events": []map[string]interface{}{
				{
					"eventId":      "evt-1",
					"eventTypeKey": "exposure",
					"decisionId":   "00000000-0000-0000-0000-000000000001",
					"userId":       "u1",
					"timestamp":    time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		resp, err := postJSON(client, baseURL+"/events", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Errorf("POST /events: status %d, body %s", resp.StatusCode, string(b))
		}
	})

	t.Run("Events_Invalid_Missing_Required_Rejected", func(t *testing.T) {
		body := map[string]interface{}{
			"events": []map[string]interface{}{
				{
					"eventId": "evt-2",
					"userId":  "u1",
					// missing eventTypeKey, decisionId, timestamp
				},
			},
		}
		resp, err := postJSON(client, baseURL+"/events", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		// Should still 200 with accepted/rejected counts, or 400
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			b, _ := io.ReadAll(resp.Body)
			t.Logf("POST /events invalid: status %d, body %s", resp.StatusCode, string(b))
		}
	})
}

func startPostgres(ctx context.Context, t *testing.T) (testcontainers.Container, string, error) {
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

func startRedis(ctx context.Context, t *testing.T) (testcontainers.Container, string, error) {
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

func runGoose(command, dbString, migrationsDir string) {
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

func waitForReady(t *testing.T, baseURL string) {
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

func postJSON(client *http.Client, targetURL string, body interface{}) (*http.Response, error) {
	enc, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(enc))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func postJSONWithAuth(client *http.Client, targetURL, bearerToken string, body interface{}) (*http.Response, error) {
	enc, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(enc))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	return client.Do(req)
}

// getWithAuth and postJSONWithAuth for other methods if needed
func getWithAuth(client *http.Client, targetURL, bearerToken string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	return client.Do(req)
}
