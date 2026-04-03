package internal_test

import (
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-approval/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func TestNewPlugin_ImplementsPluginProvider(t *testing.T) {
	var _ sdk.PluginProvider = internal.NewPlugin()
}

func TestNewPlugin_ImplementsModuleProvider(t *testing.T) {
	p := internal.NewPlugin()
	if _, ok := p.(sdk.ModuleProvider); !ok {
		t.Error("plugin does not implement ModuleProvider")
	}
}

func TestNewPlugin_ImplementsStepProvider(t *testing.T) {
	p := internal.NewPlugin()
	if _, ok := p.(sdk.StepProvider); !ok {
		t.Error("plugin does not implement StepProvider")
	}
}

func TestManifest_HasRequiredFields(t *testing.T) {
	m := internal.Manifest
	if m.Name == "" {
		t.Error("manifest Name is empty")
	}
	if m.Version == "" {
		t.Error("manifest Version is empty")
	}
	if m.Description == "" {
		t.Error("manifest Description is empty")
	}
}

func TestModuleTypes(t *testing.T) {
	p := internal.NewPlugin().(sdk.ModuleProvider)
	types := p.ModuleTypes()
	if len(types) != 1 || types[0] != "approval.engine" {
		t.Errorf("expected [approval.engine], got %v", types)
	}
}

func TestStepTypes(t *testing.T) {
	p := internal.NewPlugin().(sdk.StepProvider)
	types := p.StepTypes()
	expected := map[string]bool{
		"step.approval_request":  true,
		"step.approval_check":    true,
		"step.approval_decide":   true,
		"step.approval_list":     true,
		"step.approval_escalate": true,
		"step.approval_wait":     true,
	}
	if len(types) != len(expected) {
		t.Errorf("expected %d step types, got %d", len(expected), len(types))
	}
	for _, st := range types {
		if !expected[st] {
			t.Errorf("unexpected step type: %s", st)
		}
	}
}

func TestCreateModule_ApprovalEngine(t *testing.T) {
	p := internal.NewPlugin().(sdk.ModuleProvider)
	m, err := p.CreateModule("approval.engine", "test-engine", nil)
	if err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}
	if err := m.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
}

func TestCreateModule_UnknownType(t *testing.T) {
	p := internal.NewPlugin().(sdk.ModuleProvider)
	_, err := p.CreateModule("unknown.type", "test", nil)
	if err == nil {
		t.Error("expected error for unknown module type")
	}
}

func TestCreateStep_UnknownType(t *testing.T) {
	p := internal.NewPlugin().(sdk.StepProvider)
	_, err := p.CreateStep("step.unknown", "test", nil)
	if err == nil {
		t.Error("expected error for unknown step type")
	}
}

func TestApprovalEngine_ImplementsServiceInvoker(t *testing.T) {
	p := internal.NewPlugin().(sdk.ModuleProvider)
	m, _ := p.CreateModule("approval.engine", "eng", nil)
	_ = m.Init()
	if _, ok := m.(sdk.ServiceInvoker); !ok {
		t.Error("approval engine does not implement ServiceInvoker")
	}
}

func TestApprovalEngine_ImplementsMessageAwareModule(t *testing.T) {
	p := internal.NewPlugin().(sdk.ModuleProvider)
	m, _ := p.CreateModule("approval.engine", "eng", nil)
	_ = m.Init()
	if _, ok := m.(sdk.MessageAwareModule); !ok {
		t.Error("approval engine does not implement MessageAwareModule")
	}
}
