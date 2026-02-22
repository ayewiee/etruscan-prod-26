package models

import "testing"

func TestExperimentStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from ExperimentStatus
		to   ExperimentStatus
		want bool
	}{
		{ExperimentStatusDraft, ExperimentStatusOnReview, true},
		{ExperimentStatusDraft, ExperimentStatusApproved, false},
		{ExperimentStatusDraft, ExperimentStatusLaunched, false},
		{ExperimentStatusOnReview, ExperimentStatusApproved, true},
		{ExperimentStatusOnReview, ExperimentStatusDraft, true},
		{ExperimentStatusOnReview, ExperimentStatusDeclined, true},
		{ExperimentStatusOnReview, ExperimentStatusLaunched, false},
		{ExperimentStatusApproved, ExperimentStatusLaunched, true},
		{ExperimentStatusApproved, ExperimentStatusDraft, false},
		{ExperimentStatusLaunched, ExperimentStatusPaused, true},
		{ExperimentStatusLaunched, ExperimentStatusFinished, true},
		{ExperimentStatusLaunched, ExperimentStatusDraft, false},
		{ExperimentStatusPaused, ExperimentStatusLaunched, true},
		{ExperimentStatusPaused, ExperimentStatusFinished, true},
		{ExperimentStatusFinished, ExperimentStatusArchived, true},
		{ExperimentStatusFinished, ExperimentStatusLaunched, false},
	}
	for _, tt := range tests {
		got := tt.from.CanTransitionTo(tt.to)
		if got != tt.want {
			t.Errorf("CanTransitionTo(%q -> %q) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestExperimentStatus_EditingAllowed(t *testing.T) {
	if !ExperimentStatusDraft.EditingAllowed() {
		t.Error("DRAFT should allow editing")
	}
	for _, s := range []ExperimentStatus{
		ExperimentStatusOnReview, ExperimentStatusApproved, ExperimentStatusLaunched,
		ExperimentStatusPaused, ExperimentStatusFinished, ExperimentStatusArchived, ExperimentStatusDeclined,
	} {
		if s.EditingAllowed() {
			t.Errorf("%q should not allow editing", s)
		}
	}
}

func TestExperimentStatus_ApprovalAllowed(t *testing.T) {
	if !ExperimentStatusOnReview.ApprovalAllowed() {
		t.Error("ON_REVIEW should allow approval")
	}
	for _, s := range []ExperimentStatus{
		ExperimentStatusDraft, ExperimentStatusApproved, ExperimentStatusLaunched,
		ExperimentStatusPaused, ExperimentStatusFinished, ExperimentStatusArchived, ExperimentStatusDeclined,
	} {
		if s.ApprovalAllowed() {
			t.Errorf("%q should not allow approval", s)
		}
	}
}
