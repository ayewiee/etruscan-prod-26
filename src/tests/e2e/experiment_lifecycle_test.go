package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RunExperimentLifecycleTests(t *testing.T, client *http.Client, baseURL, token string) {
	t.Run("B3_1_Experiment_Draft_To_InReview", func(t *testing.T) {
		flagKey := "e2e_draft_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "draft to review", 100)
		SendOnReview(t, client, baseURL, token, expID)
		status := GetExperimentStatus(t, client, baseURL, token, expID)
		assert.Equal(t, "ON_REVIEW", status, "after sendOnReview")
	})

	t.Run("B3_2_Experiment_InReview_To_Approved", func(t *testing.T) {
		flagKey := "e2e_approve_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "approve flow", 100)
		CreateApproverGroupAndAssignToAdmin(t, client, baseURL, token)
		SendOnReview(t, client, baseURL, token, expID)
		ApproveExperiment(t, client, baseURL, token, expID)
		status := GetExperimentStatus(t, client, baseURL, token, expID)
		assert.Equal(t, "APPROVED", status, "after approve")
	})

	t.Run("B3_3_Launch_Blocked_Without_Approvals", func(t *testing.T) {
		flagKey := "e2e_nolaunch_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "no approval", 100)
		SendOnReview(t, client, baseURL, token, expID)
		resp, err := PostJSONWithAuth(client, baseURL+"/experiments/"+expID+"/launch", token, map[string]string{})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusOK, resp.StatusCode, "launch without approval should be blocked")
	})

	t.Run("B3_4_Invalid_Status_Transition_Blocked", func(t *testing.T) {
		flagKey := "e2e_invalid_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "invalid transition", 100)
		resp, err := PostJSONWithAuth(client, baseURL+"/experiments/"+expID+"/launch", token, map[string]string{})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusOK, resp.StatusCode, "launch from DRAFT should be blocked")
	})

	t.Run("B3_5_Viewer_Cannot_Approve", func(t *testing.T) {
		flagKey := "e2e_viewer_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "viewer cannot approve", 100)
		SendOnReview(t, client, baseURL, token, expID)
		viewerToken := CreateViewerAndLogin(t, client, baseURL, token)
		resp, err := PostJSONWithAuth(client, baseURL+"/experiments/"+expID+"/approve", viewerToken, map[string]string{})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, http.StatusOK, resp.StatusCode, "viewer should not be able to approve")
	})
}
