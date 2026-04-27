package internal

import (
	"context"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-approval/internal/contracts"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func TestTypedApprovalWaitTimeoutPreservesRequestID(t *testing.T) {
	engine := NewApprovalEngine(nil)
	if err := engine.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	created, err := engine.store.Create(&ApprovalRequest{
		Title:     "wait timeout",
		Approvers: []string{"alice"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	result, err := typedApprovalWait(engine)(context.Background(), sdk.TypedStepRequest[*contracts.ApprovalWaitConfig, *contracts.ApprovalWaitArgs]{
		Config: &contracts.ApprovalWaitConfig{TimeoutSeconds: 1, PollIntervalMs: 1},
		Input:  &contracts.ApprovalWaitArgs{RequestId: created.ID},
	})
	if err != nil {
		t.Fatalf("typedApprovalWait: %v", err)
	}
	if result == nil || result.Output == nil {
		t.Fatal("expected typed wait output")
	}
	if result.Output.Id != created.ID {
		t.Fatalf("output id = %q, want %q", result.Output.Id, created.ID)
	}
	if !result.Output.TimedOut {
		t.Fatal("expected timed_out output")
	}
	if !result.StopPipeline {
		t.Fatal("expected wait timeout to stop pipeline")
	}
}

func TestPluginResolveEngineByNameUsesTypedModuleRegistry(t *testing.T) {
	p := NewPlugin().(*plugin)
	first, err := p.CreateTypedModule("approval.engine", "first", nil)
	if err != nil {
		t.Fatalf("CreateTypedModule first: %v", err)
	}
	second, err := p.CreateTypedModule("approval.engine", "second", nil)
	if err != nil {
		t.Fatalf("CreateTypedModule second: %v", err)
	}
	if got := p.resolveEngineByName("first"); got != first.(*sdk.TypedModuleInstance[*contracts.ApprovalEngineConfig]).ModuleInstance {
		t.Fatal("first engine did not resolve to the first typed module")
	}
	if got := p.resolveEngineByName("second"); got != second.(*sdk.TypedModuleInstance[*contracts.ApprovalEngineConfig]).ModuleInstance {
		t.Fatal("second engine did not resolve to the second typed module")
	}
	if got := p.resolveEngineByName("missing"); got != nil {
		t.Fatal("missing engine resolved unexpectedly")
	}
}
