package internal_test

import (
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-approval/internal"
	"github.com/GoCodeAlone/workflow-plugin-approval/internal/contracts"
	pb "github.com/GoCodeAlone/workflow/plugin/external/proto"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"google.golang.org/protobuf/types/known/anypb"
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

func TestNewPlugin_ImplementsStrictContractProviders(t *testing.T) {
	p := internal.NewPlugin()
	if _, ok := p.(sdk.TypedModuleProvider); !ok {
		t.Error("plugin does not implement TypedModuleProvider")
	}
	if _, ok := p.(sdk.TypedStepProvider); !ok {
		t.Error("plugin does not implement TypedStepProvider")
	}
	if _, ok := p.(sdk.ContractProvider); !ok {
		t.Error("plugin does not implement ContractProvider")
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

func TestContractRegistry_DeclaresStrictModuleStepAndServiceContracts(t *testing.T) {
	p := internal.NewPlugin().(sdk.ContractProvider)
	registry := p.ContractRegistry()
	if registry == nil {
		t.Fatal("expected contract registry")
	}
	if registry.FileDescriptorSet == nil || len(registry.FileDescriptorSet.File) == 0 {
		t.Fatal("expected file descriptor set")
	}
	contractsByKey := make(map[string]*pb.ContractDescriptor)
	for _, contract := range registry.Contracts {
		switch contract.Kind {
		case pb.ContractKind_CONTRACT_KIND_MODULE:
			contractsByKey["module:"+contract.ModuleType] = contract
		case pb.ContractKind_CONTRACT_KIND_STEP:
			contractsByKey["step:"+contract.StepType] = contract
		case pb.ContractKind_CONTRACT_KIND_SERVICE:
			contractsByKey["service:"+contract.Method] = contract
		}
	}
	for _, key := range []string{
		"module:approval.engine",
		"step:step.approval_request",
		"step:step.approval_check",
		"step:step.approval_decide",
		"step:step.approval_list",
		"step:step.approval_escalate",
		"step:step.approval_wait",
		"service:create",
		"service:get",
		"service:decide",
		"service:list",
		"service:escalate",
		"service:check_expiry",
	} {
		contract, ok := contractsByKey[key]
		if !ok {
			t.Fatalf("missing contract %s", key)
		}
		if contract.Mode != pb.ContractMode_CONTRACT_MODE_STRICT_PROTO {
			t.Fatalf("%s mode = %s, want strict proto", key, contract.Mode)
		}
	}
}

func TestTypedModule_InvokeTypedServiceCreate(t *testing.T) {
	p := internal.NewPlugin().(sdk.TypedModuleProvider)
	module, err := p.CreateTypedModule("approval.engine", "eng", nil)
	if err != nil {
		t.Fatalf("CreateTypedModule: %v", err)
	}
	if err := module.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	invoker, ok := module.(sdk.TypedServiceInvoker)
	if !ok {
		t.Fatal("typed module does not implement TypedServiceInvoker")
	}
	input, err := anypb.New(&contracts.ApprovalCreateArgs{
		Title:             "Deploy v4",
		Approvers:         []string{"alice"},
		RequiredApprovals: 1,
	})
	if err != nil {
		t.Fatalf("pack typed input: %v", err)
	}
	output, err := invoker.InvokeTypedMethod("create", input)
	if err != nil {
		t.Fatalf("InvokeTypedMethod: %v", err)
	}
	var request contracts.ApprovalRequest
	if err := output.UnmarshalTo(&request); err != nil {
		t.Fatalf("unpack typed output: %v", err)
	}
	if request.Status != "pending" {
		t.Fatalf("status = %q, want pending", request.Status)
	}
	if request.Title != "Deploy v4" {
		t.Fatalf("title = %q, want Deploy v4", request.Title)
	}
}

func TestCreateTypedStep_WithNilConfigUsesStrictDefaultConfig(t *testing.T) {
	provider := internal.NewPlugin()
	mp := provider.(sdk.TypedModuleProvider)
	sp := provider.(sdk.TypedStepProvider)
	if _, err := mp.CreateTypedModule("approval.engine", "eng", nil); err != nil {
		t.Fatalf("CreateTypedModule: %v", err)
	}
	if _, err := sp.CreateTypedStep("step.approval_request", "request", nil); err != nil {
		t.Fatalf("CreateTypedStep: %v", err)
	}
}

func TestCreateTypedStep_UsesNamedEngineFromTypedConfig(t *testing.T) {
	provider := internal.NewPlugin()
	mp := provider.(sdk.TypedModuleProvider)
	sp := provider.(sdk.TypedStepProvider)
	if _, err := mp.CreateTypedModule("approval.engine", "first", nil); err != nil {
		t.Fatalf("CreateTypedModule first: %v", err)
	}
	if _, err := mp.CreateTypedModule("approval.engine", "second", nil); err != nil {
		t.Fatalf("CreateTypedModule second: %v", err)
	}
	config, err := anypb.New(&contracts.ApprovalCreateConfig{Engine: "second"})
	if err != nil {
		t.Fatalf("pack config: %v", err)
	}
	if _, err := sp.CreateTypedStep("step.approval_request", "request", config); err != nil {
		t.Fatalf("CreateTypedStep: %v", err)
	}
	missing, err := anypb.New(&contracts.ApprovalCreateConfig{Engine: "missing"})
	if err != nil {
		t.Fatalf("pack missing config: %v", err)
	}
	if _, err := sp.CreateTypedStep("step.approval_request", "request", missing); err == nil {
		t.Fatal("expected missing named engine error")
	}
}
