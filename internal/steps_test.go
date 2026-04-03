package internal_test

import (
	"context"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-approval/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// helper to create a plugin with a ready engine and all step types.
func setupPlugin(t *testing.T) (sdk.StepProvider, sdk.ModuleProvider) {
	t.Helper()
	p := internal.NewPlugin()
	mp := p.(sdk.ModuleProvider)
	sp := p.(sdk.StepProvider)
	m, err := mp.CreateModule("approval.engine", "eng", nil)
	if err != nil {
		t.Fatalf("CreateModule: %v", err)
	}
	if err := m.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return sp, mp
}

func TestStepApprovalRequest(t *testing.T) {
	sp, _ := setupPlugin(t)
	step, err := sp.CreateStep("step.approval_request", "req1", nil)
	if err != nil {
		t.Fatalf("CreateStep: %v", err)
	}
	result, err := step.Execute(context.Background(), nil, nil, map[string]any{
		"title":              "Deploy v3",
		"approvers":         []any{"alice", "bob"},
		"required_approvals": 2,
	}, nil, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Output["status"] != "pending" {
		t.Errorf("expected pending, got %v", result.Output["status"])
	}
	if result.Output["id"] == nil || result.Output["id"] == "" {
		t.Error("expected non-empty id")
	}
}

func TestStepApprovalCheck(t *testing.T) {
	sp, _ := setupPlugin(t)
	// First create a request
	reqStep, _ := sp.CreateStep("step.approval_request", "req1", nil)
	res, _ := reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "Check Test",
		"approvers": []any{"alice"},
	}, nil, nil)
	reqID := res.Output["id"].(string)

	// Now check it
	checkStep, _ := sp.CreateStep("step.approval_check", "chk1", nil)
	result, err := checkStep.Execute(context.Background(), nil, nil, map[string]any{
		"request_id": reqID,
	}, nil, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Output["status"] != "pending" {
		t.Errorf("expected pending, got %v", result.Output["status"])
	}
}

func TestStepApprovalDecide(t *testing.T) {
	sp, _ := setupPlugin(t)
	reqStep, _ := sp.CreateStep("step.approval_request", "req1", nil)
	res, _ := reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "Decide Test",
		"approvers": []any{"alice"},
	}, nil, nil)
	reqID := res.Output["id"].(string)

	decideStep, _ := sp.CreateStep("step.approval_decide", "dec1", nil)
	result, err := decideStep.Execute(context.Background(), nil, nil, map[string]any{
		"request_id": reqID,
		"actor":      "alice",
		"decision":   "approve",
		"comment":    "LGTM",
	}, nil, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Output["status"] != "approved" {
		t.Errorf("expected approved, got %v", result.Output["status"])
	}
}

func TestStepApprovalDecide_InvalidDecision(t *testing.T) {
	sp, _ := setupPlugin(t)
	reqStep, _ := sp.CreateStep("step.approval_request", "req1", nil)
	res, _ := reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "Invalid Decision",
		"approvers": []any{"alice"},
	}, nil, nil)
	reqID := res.Output["id"].(string)

	decideStep, _ := sp.CreateStep("step.approval_decide", "dec1", nil)
	_, err := decideStep.Execute(context.Background(), nil, nil, map[string]any{
		"request_id": reqID,
		"actor":      "alice",
		"decision":   "maybe",
	}, nil, nil)
	if err == nil {
		t.Error("expected error for invalid decision")
	}
}

func TestStepApprovalList(t *testing.T) {
	sp, _ := setupPlugin(t)
	reqStep, _ := sp.CreateStep("step.approval_request", "req1", nil)
	reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "List Test 1",
		"approvers": []any{"alice"},
	}, nil, nil)
	reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "List Test 2",
		"approvers": []any{"bob"},
	}, nil, nil)

	listStep, _ := sp.CreateStep("step.approval_list", "list1", nil)
	result, err := listStep.Execute(context.Background(), nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	count, _ := result.Output["count"].(int)
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestStepApprovalList_FilterByApprover(t *testing.T) {
	sp, _ := setupPlugin(t)
	reqStep, _ := sp.CreateStep("step.approval_request", "req1", nil)
	reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "Alice req",
		"approvers": []any{"alice"},
	}, nil, nil)
	reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "Bob req",
		"approvers": []any{"bob"},
	}, nil, nil)

	listStep, _ := sp.CreateStep("step.approval_list", "list1", nil)
	result, _ := listStep.Execute(context.Background(), nil, nil, map[string]any{
		"approver": "alice",
	}, nil, nil)
	count, _ := result.Output["count"].(int)
	if count != 1 {
		t.Errorf("expected 1 for alice, got %d", count)
	}
}

func TestStepApprovalEscalate(t *testing.T) {
	sp, _ := setupPlugin(t)
	reqStep, _ := sp.CreateStep("step.approval_request", "req1", nil)
	res, _ := reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "Escalate Test",
		"approvers": []any{"alice"},
	}, nil, nil)
	reqID := res.Output["id"].(string)

	escStep, _ := sp.CreateStep("step.approval_escalate", "esc1", nil)
	result, err := escStep.Execute(context.Background(), nil, nil, map[string]any{
		"request_id":    reqID,
		"new_approvers": []any{"manager1"},
	}, nil, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Output["status"] != "escalated" {
		t.Errorf("expected escalated, got %v", result.Output["status"])
	}
}

func TestStepApprovalWait_AlreadyDecided(t *testing.T) {
	sp, _ := setupPlugin(t)
	reqStep, _ := sp.CreateStep("step.approval_request", "req1", nil)
	res, _ := reqStep.Execute(context.Background(), nil, nil, map[string]any{
		"title":     "Wait Test",
		"approvers": []any{"alice"},
	}, nil, nil)
	reqID := res.Output["id"].(string)

	// Decide first
	decideStep, _ := sp.CreateStep("step.approval_decide", "dec1", nil)
	decideStep.Execute(context.Background(), nil, nil, map[string]any{
		"request_id": reqID,
		"actor":      "alice",
		"decision":   "approve",
	}, nil, nil)

	// Wait should return immediately
	waitStep, _ := sp.CreateStep("step.approval_wait", "wait1", nil)
	result, err := waitStep.Execute(context.Background(), nil, nil, map[string]any{
		"request_id": reqID,
	}, nil, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Output["status"] != "approved" {
		t.Errorf("expected approved, got %v", result.Output["status"])
	}
}

func TestStepApprovalWait_MissingRequestID(t *testing.T) {
	sp, _ := setupPlugin(t)
	waitStep, _ := sp.CreateStep("step.approval_wait", "wait1", nil)
	_, err := waitStep.Execute(context.Background(), nil, nil, nil, nil, nil)
	if err == nil {
		t.Error("expected error for missing request_id")
	}
}

func TestStepApprovalRequest_MergesConfigAndCurrent(t *testing.T) {
	sp, _ := setupPlugin(t)
	step, _ := sp.CreateStep("step.approval_request", "req1", nil)
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{"title": "from current"},
		nil,
		map[string]any{"approvers": []any{"alice"}},
	)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Output["title"] != "from current" {
		t.Errorf("expected title from current, got %v", result.Output["title"])
	}
}
