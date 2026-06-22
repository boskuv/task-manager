package domain

import "testing"

func TestTeamRoleIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		role  TeamRole
		valid bool
	}{
		{TeamRoleOwner, true},
		{TeamRoleAdmin, true},
		{TeamRoleMember, true},
		{TeamRole("guest"), false},
		{TeamRole(""), false},
	}

	for _, tt := range tests {
		if got := tt.role.IsValid(); got != tt.valid {
			t.Errorf("TeamRole(%q).IsValid() = %v, want %v", tt.role, got, tt.valid)
		}
	}
}

func TestTeamRoleCanInvite(t *testing.T) {
	t.Parallel()

	if !TeamRoleOwner.CanInvite() {
		t.Error("owner should be able to invite")
	}
	if !TeamRoleAdmin.CanInvite() {
		t.Error("admin should be able to invite")
	}
	if TeamRoleMember.CanInvite() {
		t.Error("member should not be able to invite")
	}
}
