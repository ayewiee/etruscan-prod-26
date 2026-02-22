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

func RunFlagsDecideTests(t *testing.T, client *http.Client, baseURL, token string) {
	t.Run("Happy_Flags_CRUD", func(t *testing.T) {
		createBody := map[string]interface{}{
			"key":          "e2e_flag_" + fmt.Sprintf("%d", time.Now().UnixNano()),
			"valueType":    "string",
			"defaultValue": "control",
		}
		resp, err := PostJSONWithAuth(client, baseURL+"/flags", token, createBody)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode, "POST /flags")
		var flag struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&flag))
		assert.NotEmpty(t, flag.ID)
		assert.NotEmpty(t, flag.Key)
	})

	t.Run("Decide_Without_Active_Experiment_Returns_Default", func(t *testing.T) {
		key := "e2e_decide_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, key, "default_val")
		dec := Decide(t, client, baseURL, "user-1", key)
		assert.Equal(t, "default_val", dec.Value, "decide without experiment")
	})

	t.Run("B1_5_HappyPath_Decide_Event_Report", func(t *testing.T) {
		flagKey := "e2e_happy_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "control")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "e2e happy path", 100)
		CreateApproverGroupAndAssignToAdmin(t, client, baseURL, token)
		SendOnReview(t, client, baseURL, token, expID)
		ApproveExperiment(t, client, baseURL, token, expID)
		LaunchExperiment(t, client, baseURL, token, expID)
		decResp := Decide(t, client, baseURL, "user-happy", flagKey)
		SendExposureEvent(t, client, baseURL, decResp.DecisionID, "user-happy")
		SendConversionEvent(t, client, baseURL, decResp.DecisionID, "user-happy")
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now().Add(time.Hour)
		GetReport(t, client, baseURL, token, expID, from, to)
		FinishWithOutcome(t, client, baseURL, token, expID, "NO_EFFECT", "e2e happy path completed")
	})

	t.Run("B2_3_Decide_Returns_Variant_When_Experiment_Applicable", func(t *testing.T) {
		flagKey := "e2e_variant_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "off")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "variant test", 100)
		CreateApproverGroupAndAssignToAdmin(t, client, baseURL, token)
		SendOnReview(t, client, baseURL, token, expID)
		ApproveExperiment(t, client, baseURL, token, expID)
		LaunchExperiment(t, client, baseURL, token, expID)
		dec := Decide(t, client, baseURL, "user-variant", flagKey)
		assert.Contains(t, []interface{}{"control", "treatment"}, dec.Value, "decide returns variant")
	})

	t.Run("B2_4_Determinism_Same_User_Same_Result", func(t *testing.T) {
		flagKey := "e2e_det_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "default")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "determinism", 100)
		CreateApproverGroupAndAssignToAdmin(t, client, baseURL, token)
		SendOnReview(t, client, baseURL, token, expID)
		ApproveExperiment(t, client, baseURL, token, expID)
		LaunchExperiment(t, client, baseURL, token, expID)
		u := "user-det-" + fmt.Sprintf("%d", time.Now().UnixNano())
		d1 := Decide(t, client, baseURL, u, flagKey)
		d2 := Decide(t, client, baseURL, u, flagKey)
		assert.Equal(t, d1.Value, d2.Value, "determinism: same user same result")
	})
}
