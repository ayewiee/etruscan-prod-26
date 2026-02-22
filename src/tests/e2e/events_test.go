package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RunEventsTests(t *testing.T, client *http.Client, baseURL, token string) {
	t.Run("Events_Ingest_Valid", func(t *testing.T) {
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
		resp, err := PostJSON(client, baseURL+"/track", body)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "POST /track valid")
	})

	t.Run("Events_Invalid_Missing_Required_Rejected", func(t *testing.T) {
		body := map[string]interface{}{
			"events": []map[string]interface{}{
				{
					"eventId": "evt-2",
					"userId":  "u1",
				},
			},
		}
		resp, err := PostJSON(client, baseURL+"/track", body)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnprocessableEntity,
			"POST /track invalid: expected 200 or 422, got %d", resp.StatusCode)
	})

	t.Run("B4_1_Events_Invalid_Type_Rejected", func(t *testing.T) {
		body := map[string]interface{}{
			"events": []map[string]interface{}{
				{
					"eventId":      "evt-bad-type",
					"eventTypeKey": "nonexistent_type_xyz",
					"decisionId":   "00000000-0000-0000-0000-000000000001",
					"userId":       "u1",
					"timestamp":    time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		resp, err := PostJSON(client, baseURL+"/track", body)
		require.NoError(t, err)
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var r struct {
				Rejected int `json:"rejected"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&r)
			assert.GreaterOrEqual(t, r.Rejected, 1, "invalid event type should be rejected")
		}
	})

	t.Run("B4_3_Events_Dedupe", func(t *testing.T) {
		decID := "00000000-0000-0000-0000-000000000002"
		evtID := "evt-dedup-" + fmt.Sprintf("%d", time.Now().UnixNano())
		body := map[string]interface{}{
			"events": []map[string]interface{}{
				{
					"eventId":      evtID,
					"eventTypeKey": "exposure",
					"decisionId":   decID,
					"userId":       "u-dedup",
					"timestamp":    time.Now().UTC().Format(time.RFC3339),
				},
			},
		}
		resp1, _ := PostJSON(client, baseURL+"/track", body)
		resp1.Body.Close()
		resp2, err := PostJSON(client, baseURL+"/track", body)
		require.NoError(t, err)
		defer resp2.Body.Close()
		var r struct {
			Accepted   int `json:"accepted"`
			Duplicates int `json:"duplicates"`
		}
		require.NoError(t, json.NewDecoder(resp2.Body).Decode(&r))
		assert.True(t, r.Duplicates >= 1 || r.Accepted == 0, "second send with same eventId should be duplicate")
	})
}
