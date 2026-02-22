package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RunReportGuardrailTests(t *testing.T, client *http.Client, baseURL, token string) {
	t.Run("B5_1_2_Guardrail_MetricKey_And_Threshold", func(t *testing.T) {
		flagKey := "e2e_guard_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperimentWithGuardrail(t, client, baseURL, token, flagID, "guardrail test")
		exp := GetExperiment(t, client, baseURL, token, expID)
		require.GreaterOrEqual(t, len(exp.Guardrails), 1, "experiment should have at least one guardrail")
		g := exp.Guardrails[0]
		assert.NotEmpty(t, g.MetricKey, "guardrail must have metricKey")
		assert.Greater(t, g.Threshold, 0.0, "guardrail must have positive threshold")
	})

	t.Run("B6_1_Report_Period_Filter", func(t *testing.T) {
		flagKey := "e2e_rpt_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "report period", 100)
		CreateApproverGroupAndAssignToAdmin(t, client, baseURL, token)
		SendOnReview(t, client, baseURL, token, expID)
		ApproveExperiment(t, client, baseURL, token, expID)
		LaunchExperiment(t, client, baseURL, token, expID)
		from := time.Now().Add(-7 * 24 * time.Hour)
		to := time.Now().Add(24 * time.Hour)
		report := GetReport(t, client, baseURL, token, expID, from, to)
		assert.NotEmpty(t, report.From, "report should have from")
		assert.NotEmpty(t, report.To, "report should have to")
	})

	t.Run("B6_2_Report_By_Variant", func(t *testing.T) {
		flagKey := "e2e_var_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "report variants", 100)
		CreateApproverGroupAndAssignToAdmin(t, client, baseURL, token)
		SendOnReview(t, client, baseURL, token, expID)
		ApproveExperiment(t, client, baseURL, token, expID)
		LaunchExperiment(t, client, baseURL, token, expID)
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now().Add(time.Hour)
		report := GetReport(t, client, baseURL, token, expID, from, to)
		assert.GreaterOrEqual(t, len(report.Variants), 2, "report should have variants breakdown")
	})

	t.Run("B6_4_5_Finish_With_Outcome_And_Comment", func(t *testing.T) {
		flagKey := "e2e_fin_" + fmt.Sprintf("%d", time.Now().UnixNano())
		CreateFlag(t, client, baseURL, token, flagKey, "x")
		flagID := GetFlagID(t, client, baseURL, token, flagKey)
		expID := CreateExperiment(t, client, baseURL, token, flagID, "finish outcome", 100)
		CreateApproverGroupAndAssignToAdmin(t, client, baseURL, token)
		SendOnReview(t, client, baseURL, token, expID)
		ApproveExperiment(t, client, baseURL, token, expID)
		LaunchExperiment(t, client, baseURL, token, expID)
		FinishWithOutcome(t, client, baseURL, token, expID, "ROLLOUT", "e2e decided to rollout")
		exp := GetExperiment(t, client, baseURL, token, expID)
		require.NotNil(t, exp.Outcome, "experiment outcome should be set")
		assert.Equal(t, "ROLLOUT", *exp.Outcome)
		require.NotNil(t, exp.OutcomeComment, "outcome comment should be saved")
		assert.Equal(t, "e2e decided to rollout", *exp.OutcomeComment)
	})
}
