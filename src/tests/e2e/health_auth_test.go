package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RunHealthAuthTests(t *testing.T, client *http.Client, baseURL, token string) {
	t.Run("Health", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "GET /health")
	})

	t.Run("Ready", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/ready")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "GET /ready")
	})

	t.Run("Unauthenticated_Protected_Returns_401", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/flags")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "GET /flags without token")
	})

	t.Run("Login_Invalid_Credentials_Returns_401", func(t *testing.T) {
		body := map[string]string{"email": AdminEmail, "password": "wrong"}
		resp, err := PostJSON(client, baseURL+"/auth/login", body)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "login wrong password")
	})

	t.Run("Login_Success", func(t *testing.T) {
		body := map[string]string{"email": AdminEmail, "password": AdminPass}
		resp, err := PostJSON(client, baseURL+"/auth/login", body)
		require.NoError(t, err)
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode, "login: %s", string(bodyBytes))
		var out struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(bodyBytes, &out))
		assert.NotEmpty(t, out.Token, "token")
	})

	t.Run("RBAC_Admin_Can_List_Users", func(t *testing.T) {
		resp, err := GetWithAuth(client, baseURL+"/admin/users", token)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "GET /admin/users as admin")
	})

	t.Run("RBAC_Viewer_Cannot_Access_Admin", func(t *testing.T) {
		createBody := map[string]interface{}{
			"username": "viewer1",
			"email":    "viewer@e2e.test",
			"password": "password",
			"role":     "VIEWER",
		}
		resp, err := PostJSONWithAuth(client, baseURL+"/admin/users", token, createBody)
		require.NoError(t, err)
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Skip("create viewer user failed (maybe email exists)")
		}
		loginResp, err := PostJSON(client, baseURL+"/auth/login", map[string]string{"email": "viewer@e2e.test", "password": "password"})
		require.NoError(t, err)
		defer loginResp.Body.Close()
		if loginResp.StatusCode != http.StatusOK {
			t.Skip("viewer login failed")
		}
		var loginOut struct {
			Token string `json:"token"`
		}
		_ = json.NewDecoder(loginResp.Body).Decode(&loginOut)
		resp2, err := GetWithAuth(client, baseURL+"/admin/users", loginOut.Token)
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp2.StatusCode, "GET /admin/users as VIEWER")
	})
}
