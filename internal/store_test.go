package internal

import (
	"testing"
	"time"
)

func TestStore_CreateAndGet(t *testing.T) {
	s := NewStore()
	req, err := s.Create(&ApprovalRequest{
		Title:     "Deploy v2",
		Approvers: []string{"alice", "bob"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if req.ID == "" {
		t.Error("expected generated ID")
	}
	if req.Status != StatusPending {
		t.Errorf("expected pending, got %s", req.Status)
	}
	if req.RequiredApprovals != 1 {
		t.Errorf("expected default required_approvals=1, got %d", req.RequiredApprovals)
	}

	got, err := s.Get(req.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "Deploy v2" {
		t.Errorf("title mismatch: %s", got.Title)
	}
}

func TestStore_CreateDuplicate(t *testing.T) {
	s := NewStore()
	_, _ = s.Create(&ApprovalRequest{ID: "dup-1", Approvers: []string{"a"}})
	_, err := s.Create(&ApprovalRequest{ID: "dup-1", Approvers: []string{"b"}})
	if err == nil {
		t.Error("expected duplicate error")
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := NewStore()
	_, err := s.Get("nonexistent")
	if err == nil {
		t.Error("expected not found error")
	}
}

func TestStore_Approve(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers:         []string{"alice"},
		RequiredApprovals: 1,
	})
	updated, err := s.UpdateDecision(req.ID, ApprovalDecision{
		Actor: "alice", Decision: "approve",
	})
	if err != nil {
		t.Fatalf("UpdateDecision: %v", err)
	}
	if updated.Status != StatusApproved {
		t.Errorf("expected approved, got %s", updated.Status)
	}
}

func TestStore_RejectWhenInsufficientApprovers(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers:         []string{"alice", "bob"},
		RequiredApprovals: 2,
	})
	updated, err := s.UpdateDecision(req.ID, ApprovalDecision{
		Actor: "alice", Decision: "reject",
	})
	if err != nil {
		t.Fatalf("UpdateDecision: %v", err)
	}
	// With 1 reject out of 2 approvers, and 2 required, it's impossible → rejected.
	if updated.Status != StatusRejected {
		t.Errorf("expected rejected, got %s", updated.Status)
	}
}

func TestStore_MultipleApprovals(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers:         []string{"alice", "bob", "carol"},
		RequiredApprovals: 2,
	})
	s.UpdateDecision(req.ID, ApprovalDecision{Actor: "alice", Decision: "approve"})
	got, _ := s.Get(req.ID)
	if got.Status != StatusPending {
		t.Errorf("expected pending after 1/2 approvals, got %s", got.Status)
	}
	s.UpdateDecision(req.ID, ApprovalDecision{Actor: "bob", Decision: "approve"})
	got, _ = s.Get(req.ID)
	if got.Status != StatusApproved {
		t.Errorf("expected approved after 2/2 approvals, got %s", got.Status)
	}
}

func TestStore_DuplicateDecision(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers:         []string{"alice"},
		RequiredApprovals: 1,
	})
	_, _ = s.UpdateDecision(req.ID, ApprovalDecision{Actor: "alice", Decision: "approve"})
	_, err := s.UpdateDecision(req.ID, ApprovalDecision{Actor: "alice", Decision: "approve"})
	if err == nil {
		t.Error("expected error on duplicate decision")
	}
}

func TestStore_UnauthorizedApprover(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers: []string{"alice"},
	})
	_, err := s.UpdateDecision(req.ID, ApprovalDecision{Actor: "eve", Decision: "approve"})
	if err == nil {
		t.Error("expected unauthorized approver error")
	}
}

func TestStore_DecisionOnTerminal(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers:         []string{"alice"},
		RequiredApprovals: 1,
	})
	s.UpdateDecision(req.ID, ApprovalDecision{Actor: "alice", Decision: "approve"})
	_, err := s.UpdateDecision(req.ID, ApprovalDecision{Actor: "alice", Decision: "reject"})
	if err == nil {
		t.Error("expected error on terminal state")
	}
}

func TestStore_Escalate(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers: []string{"alice"},
	})
	updated, err := s.Escalate(req.ID, []string{"manager1", "manager2"})
	if err != nil {
		t.Fatalf("Escalate: %v", err)
	}
	if updated.Status != StatusEscalated {
		t.Errorf("expected escalated, got %s", updated.Status)
	}
	if len(updated.Approvers) != 2 {
		t.Errorf("expected 2 approvers, got %d", len(updated.Approvers))
	}
}

func TestStore_EscalateTerminal(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers:         []string{"alice"},
		RequiredApprovals: 1,
	})
	s.UpdateDecision(req.ID, ApprovalDecision{Actor: "alice", Decision: "approve"})
	_, err := s.Escalate(req.ID, []string{"manager"})
	if err == nil {
		t.Error("expected error escalating terminal request")
	}
}

func TestStore_EscalateEmptyApprovers(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers: []string{"alice"},
	})
	_, err := s.Escalate(req.ID, nil)
	if err == nil {
		t.Error("expected error for empty approvers")
	}
}

func TestStore_CheckExpiry(t *testing.T) {
	s := NewStore()
	past := time.Now().Add(-1 * time.Hour)
	s.Create(&ApprovalRequest{
		ID:        "exp-1",
		Approvers: []string{"alice"},
		ExpiresAt: past,
	})
	s.Create(&ApprovalRequest{
		ID:        "noexp-1",
		Approvers: []string{"bob"},
	})
	expired := s.CheckExpiry(time.Now())
	if len(expired) != 1 || expired[0] != "exp-1" {
		t.Errorf("expected [exp-1], got %v", expired)
	}
	got, _ := s.Get("exp-1")
	if got.Status != StatusExpired {
		t.Errorf("expected expired, got %s", got.Status)
	}
}

func TestStore_List(t *testing.T) {
	s := NewStore()
	s.Create(&ApprovalRequest{ID: "r1", Approvers: []string{"alice", "bob"}})
	s.Create(&ApprovalRequest{ID: "r2", Approvers: []string{"carol"}})
	s.Create(&ApprovalRequest{ID: "r3", Approvers: []string{"alice"}})

	all := s.List("")
	if len(all) != 3 {
		t.Errorf("expected 3, got %d", len(all))
	}

	alice := s.List("alice")
	if len(alice) != 2 {
		t.Errorf("expected 2 for alice, got %d", len(alice))
	}
}

func TestStore_ApproveAfterEscalation(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers:         []string{"alice"},
		RequiredApprovals: 1,
	})
	s.Escalate(req.ID, []string{"manager"})
	updated, err := s.UpdateDecision(req.ID, ApprovalDecision{Actor: "manager", Decision: "approve"})
	if err != nil {
		t.Fatalf("decision after escalation: %v", err)
	}
	if updated.Status != StatusApproved {
		t.Errorf("expected approved after escalation, got %s", updated.Status)
	}
}

func TestStore_EscalationClearsStaleDecisions(t *testing.T) {
	s := NewStore()
	req, _ := s.Create(&ApprovalRequest{
		Approvers:         []string{"alice", "bob", "carol"},
		RequiredApprovals: 2,
	})
	// alice approves before escalation
	s.UpdateDecision(req.ID, ApprovalDecision{Actor: "alice", Decision: "approve"})

	// escalate to new approvers — alice's decision is now stale
	s.Escalate(req.ID, []string{"manager1", "manager2"})

	// manager1 approves — only 1 of 2 new approvers, should still be pending
	updated, err := s.UpdateDecision(req.ID, ApprovalDecision{Actor: "manager1", Decision: "approve"})
	if err != nil {
		t.Fatalf("UpdateDecision after escalation: %v", err)
	}
	if updated.Status != StatusEscalated {
		t.Errorf("expected escalated (1 of 2 new approvers), got %s", updated.Status)
	}

	// manager2 approves — now 2 of 2, should be approved
	updated, err = s.UpdateDecision(req.ID, ApprovalDecision{Actor: "manager2", Decision: "approve"})
	if err != nil {
		t.Fatalf("UpdateDecision: %v", err)
	}
	if updated.Status != StatusApproved {
		t.Errorf("expected approved after 2/2 new approvers, got %s", updated.Status)
	}
}

func TestApprovalStatus_IsTerminal(t *testing.T) {
	cases := []struct {
		status   ApprovalStatus
		terminal bool
	}{
		{StatusPending, false},
		{StatusEscalated, false},
		{StatusApproved, true},
		{StatusRejected, true},
		{StatusExpired, true},
	}
	for _, tc := range cases {
		if tc.status.IsTerminal() != tc.terminal {
			t.Errorf("IsTerminal(%s) = %v, want %v", tc.status, tc.status.IsTerminal(), tc.terminal)
		}
	}
}
