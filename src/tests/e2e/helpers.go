package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func CreateFlag(t *testing.T, client *http.Client, baseURL, token, key, defaultValue string) {
	body := map[string]interface{}{
		"key":          key,
		"valueType":    "string",
		"defaultValue": defaultValue,
	}
	resp, err := PostJSONWithAuth(client, baseURL+"/flags", token, body)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "POST /flags: %s", string(bodyBytes))
}

func GetFlagID(t *testing.T, client *http.Client, baseURL, token, key string) string {
	resp, err := GetWithAuth(client, baseURL+"/flags", token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "GET /flags")
	var list []struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	for _, f := range list {
		if f.Key == key {
			return f.ID
		}
	}
	t.Fatalf("flag %s not found", key)
	return ""
}

func CreateExperiment(t *testing.T, client *http.Client, baseURL, token, flagID, name string, audiencePct int) string {
	ctrl := true
	body := map[string]interface{}{
		"flagId":             flagID,
		"name":               name,
		"audiencePercentage": audiencePct,
		"variants": []map[string]interface{}{
			{"name": "control", "value": "control", "weight": 50, "isControl": &ctrl},
			{"name": "treatment", "value": "treatment", "weight": 50, "isControl": false},
		},
		"metricKeys": []string{"exposure_count", "conversion_count"},
	}
	resp, err := PostJSONWithAuth(client, baseURL+"/experiments", token, body)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "POST /experiments: %s", string(bodyBytes))
	var exp struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &exp))
	return exp.ID
}

func CreateExperimentWithGuardrail(t *testing.T, client *http.Client, baseURL, token, flagID, name string) string {
	ctrl := true
	body := map[string]interface{}{
		"flagId":             flagID,
		"name":               name,
		"audiencePercentage": 100,
		"variants": []map[string]interface{}{
			{"name": "control", "value": `"control"`, "weight": 100, "isControl": &ctrl},
		},
		"metricKeys": []string{"exposure_count"},
		"guardrails": []map[string]interface{}{
			{"metricKey": "error_count", "threshold": 10, "thresholdDirection": "upper", "windowSeconds": 3600, "action": "pause"},
		},
	}
	resp, err := PostJSONWithAuth(client, baseURL+"/experiments", token, body)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "POST /experiments with guardrail: %s", string(bodyBytes))
	var exp struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &exp))
	return exp.ID
}

// CreateApproverGroupAndAssignToAdmin creates an approver group, adds the seed admin to it,
// and updates the admin user with that approver group. Do not send "password" in the update
// (API validates min=8); the server keeps the existing password hash.
func CreateApproverGroupAndAssignToAdmin(t *testing.T, client *http.Client, baseURL, token string) string {
	resp, err := PostJSONWithAuth(client, baseURL+"/admin/approverGroups", token, map[string]string{"name": "e2e approvers"})
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "POST /admin/approverGroups: %s", string(bodyBytes))
	var ag struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &ag))

	adminID := "00000000-0000-0000-0000-000000000001"
	resp2, err := PostJSONWithAuth(client, baseURL+"/admin/approverGroups/"+ag.ID+"/members/add", token, map[string]interface{}{"users": []string{adminID}})
	require.NoError(t, err)
	resp2.Body.Close()
	require.Equal(t, http.StatusOK, resp2.StatusCode, "add members to approver group")

	// Update admin user: set approverGroup and minApprovals. Omit password so validation passes (min=8).
	updateBody := map[string]interface{}{
		"email":         AdminEmail,
		"username":      "admin",
		"role":          "ADMIN",
		"approverGroup": ag.ID,
		"minApprovals":  1,
	}
	resp3, err := PutJSONWithAuth(client, baseURL+"/admin/users/"+adminID, token, updateBody)
	require.NoError(t, err)
	defer resp3.Body.Close()
	bodyBytes3, _ := io.ReadAll(resp3.Body)
	require.Equal(t, http.StatusOK, resp3.StatusCode,
		"PUT /admin/users (assign approver group): status=%d body=%s", resp3.StatusCode, string(bodyBytes3))
	return ag.ID
}

func SendOnReview(t *testing.T, client *http.Client, baseURL, token, expID string) {
	resp, err := PostJSONWithAuth(client, baseURL+"/experiments/"+expID+"/sendOnReview", token, map[string]string{})
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST sendOnReview: %s", string(bodyBytes))
}

func ApproveExperiment(t *testing.T, client *http.Client, baseURL, token, expID string) {
	resp, err := PostJSONWithAuth(client, baseURL+"/experiments/"+expID+"/approve", token, map[string]string{})
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST approve: %s", string(bodyBytes))
}

func LaunchExperiment(t *testing.T, client *http.Client, baseURL, token, expID string) {
	resp, err := PostJSONWithAuth(client, baseURL+"/experiments/"+expID+"/launch", token, map[string]string{})
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST launch: %s", string(bodyBytes))
}

type DecideResult struct {
	DecisionID string
	Value      interface{}
}

func Decide(t *testing.T, client *http.Client, baseURL, userID, flagKey string) DecideResult {
	body := map[string]interface{}{"userId": userID, "flagKey": flagKey}
	resp, err := PostJSON(client, baseURL+"/decide", body)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST /decide: %s", string(bodyBytes))
	var dec struct {
		DecisionID string      `json:"decisionId"`
		Value      interface{} `json:"value"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &dec))
	return DecideResult{DecisionID: dec.DecisionID, Value: dec.Value}
}

func SendExposureEvent(t *testing.T, client *http.Client, baseURL, decisionID, userID string) {
	body := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"eventId":      "evt-exp-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				"eventTypeKey": "exposure",
				"decisionId":   decisionID,
				"userId":       userID,
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
	resp, err := PostJSON(client, baseURL+"/events", body)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST /events exposure: %s", string(bodyBytes))
}

func SendConversionEvent(t *testing.T, client *http.Client, baseURL, decisionID, userID string) {
	body := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"eventId":      "evt-conv-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				"eventTypeKey": "conversion",
				"decisionId":   decisionID,
				"userId":       userID,
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
	resp, err := PostJSON(client, baseURL+"/events", body)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST /events conversion: %s", string(bodyBytes))
}

type ReportResult struct {
	From     string
	To       string
	Variants []struct {
		VariantID   string             `json:"variantId"`
		VariantName string             `json:"variantName"`
		Metrics     map[string]float64 `json:"metrics"`
	} `json:"variants"`
}

func GetReport(t *testing.T, client *http.Client, baseURL, token, expID string, from, to time.Time) ReportResult {
	url := baseURL + "/experiments/" + expID + "/report?from=" + from.UTC().Format(time.RFC3339) + "&to=" + to.UTC().Format(time.RFC3339)
	resp, err := GetWithAuth(client, url, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "GET report: %s", string(bodyBytes))
	var r ReportResult
	require.NoError(t, json.Unmarshal(bodyBytes, &r))
	return r
}

func FinishWithOutcome(t *testing.T, client *http.Client, baseURL, token, expID, outcome, comment string) {
	body := map[string]string{"outcome": outcome, "comment": comment}
	resp, err := PostJSONWithAuth(client, baseURL+"/experiments/"+expID+"/finish", token, body)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST finish: %s", string(bodyBytes))
}

func GetExperimentStatus(t *testing.T, client *http.Client, baseURL, token, expID string) string {
	resp, err := GetWithAuth(client, baseURL+"/experiments/"+expID, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "GET experiment")
	var exp struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&exp))
	return exp.Status
}

type ExperimentResult struct {
	Status         string
	Outcome        *string
	OutcomeComment *string
	Guardrails     []struct {
		MetricKey string  `json:"metricKey"`
		Threshold float64 `json:"threshold"`
	} `json:"guardrails"`
}

func GetExperiment(t *testing.T, client *http.Client, baseURL, token, expID string) ExperimentResult {
	resp, err := GetWithAuth(client, baseURL+"/experiments/"+expID, token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "GET experiment")
	var exp ExperimentResult
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&exp))
	return exp
}

func CreateViewerAndLogin(t *testing.T, client *http.Client, baseURL, token string) string {
	createBody := map[string]interface{}{
		"username": "viewer2",
		"email":    "viewer2@e2e.test",
		"password": "password1",
		"role":     "VIEWER",
	}
	resp, err := PostJSONWithAuth(client, baseURL+"/admin/users", token, createBody)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create viewer")
	loginResp, err := PostJSON(client, baseURL+"/auth/login", map[string]string{"email": "viewer2@e2e.test", "password": "password1"})
	require.NoError(t, err)
	defer loginResp.Body.Close()
	require.Equal(t, http.StatusOK, loginResp.StatusCode, "viewer login")
	var out struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(loginResp.Body).Decode(&out))
	return out.Token
}
