package models

import "testing"

func TestUserRole_CanManageFlags(t *testing.T) {
	if !UserRoleAdmin.CanManageFlags() {
		t.Error("ADMIN should manage flags")
	}
	if !UserRoleExperimenter.CanManageFlags() {
		t.Error("EXPERIMENTER should manage flags")
	}
	if UserRoleViewer.CanManageFlags() {
		t.Error("VIEWER should not manage flags")
	}
	if UserRoleApprover.CanManageFlags() {
		t.Error("APPROVER should not manage flags")
	}
}

func TestUserRole_CanManageEventTypes(t *testing.T) {
	if !UserRoleAdmin.CanManageEventTypes() {
		t.Error("ADMIN should manage event types")
	}
	if !UserRoleExperimenter.CanManageEventTypes() {
		t.Error("EXPERIMENTER should manage event types")
	}
	if UserRoleViewer.CanManageEventTypes() {
		t.Error("VIEWER should not manage event types")
	}
	if UserRoleApprover.CanManageEventTypes() {
		t.Error("APPROVER should not manage event types")
	}
}

func TestUserRole_CanManageExperiments(t *testing.T) {
	if !UserRoleAdmin.CanManageExperiments() {
		t.Error("ADMIN should manage experiments")
	}
	if !UserRoleExperimenter.CanManageExperiments() {
		t.Error("EXPERIMENTER should manage experiments")
	}
	if UserRoleViewer.CanManageExperiments() {
		t.Error("VIEWER should not manage experiments")
	}
	if UserRoleApprover.CanManageExperiments() {
		t.Error("APPROVER should not manage experiments")
	}
}

func TestUserRole_CanApprove(t *testing.T) {
	if !UserRoleAdmin.CanApprove() {
		t.Error("ADMIN should approve")
	}
	if !UserRoleApprover.CanApprove() {
		t.Error("APPROVER should approve")
	}
	if UserRoleViewer.CanApprove() {
		t.Error("VIEWER should not approve")
	}
	if UserRoleExperimenter.CanApprove() {
		t.Error("EXPERIMENTER should not approve")
	}
}
