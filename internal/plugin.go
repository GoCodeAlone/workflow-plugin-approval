package internal

import (
	"fmt"
	"sync"

	"github.com/GoCodeAlone/workflow-plugin-approval/internal/contracts"
	pb "github.com/GoCodeAlone/workflow/plugin/external/proto"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
)

// Version is set at build time via -ldflags
// "-X github.com/GoCodeAlone/workflow-plugin-approval/internal.Version=X.Y.Z"
var Version = "dev"

// Manifest returns the plugin metadata used by the workflow engine for
// discovery and capability negotiation.
var Manifest = sdk.PluginManifest{
	Name:        "workflow-plugin-approval",
	Version:     Version,
	Description: "Human-in-the-loop approval workflows for the workflow engine",
	Author:      "GoCodeAlone",
}

// plugin implements PluginProvider, ModuleProvider, and StepProvider.
type plugin struct {
	mu      sync.RWMutex
	engines map[string]*approvalEngine
}

// NewPlugin creates a new plugin instance.
func NewPlugin() sdk.PluginProvider {
	return &plugin{engines: make(map[string]*approvalEngine)}
}

func (p *plugin) Manifest() sdk.PluginManifest { return Manifest }

// --- ModuleProvider ---

func (p *plugin) ModuleTypes() []string {
	return []string{"approval.engine"}
}

func (p *plugin) TypedModuleTypes() []string {
	return []string{"approval.engine"}
}

func (p *plugin) CreateModule(typeName, name string, config map[string]any) (sdk.ModuleInstance, error) {
	if typeName != "approval.engine" {
		return nil, fmt.Errorf("unknown module type: %s", typeName)
	}
	engine := NewApprovalEngine(config)
	p.mu.Lock()
	p.engines[name] = engine
	p.mu.Unlock()
	return engine, nil
}

func (p *plugin) CreateTypedModule(typeName, name string, config *anypb.Any) (sdk.ModuleInstance, error) {
	factory := sdk.NewTypedModuleFactory(
		"approval.engine",
		&contracts.ApprovalEngineConfig{},
		func(name string, _ *contracts.ApprovalEngineConfig) (sdk.ModuleInstance, error) {
			engine := NewApprovalEngine(nil)
			p.mu.Lock()
			p.engines[name] = engine
			p.mu.Unlock()
			return engine, nil
		},
	)
	return factory.CreateTypedModule(typeName, name, config)
}

// --- StepProvider ---

var stepTypeNames = []string{
	"step.approval_request",
	"step.approval_check",
	"step.approval_decide",
	"step.approval_list",
	"step.approval_escalate",
	"step.approval_wait",
}

func (p *plugin) StepTypes() []string { return stepTypeNames }

func (p *plugin) TypedStepTypes() []string { return stepTypeNames }

func (p *plugin) CreateStep(typeName, _ string, config map[string]any) (sdk.StepInstance, error) {
	engine := p.resolveEngine(config)
	if engine == nil {
		return nil, fmt.Errorf("no approval.engine module found; ensure one is defined in config")
	}
	switch typeName {
	case "step.approval_request":
		return &stepApprovalRequest{engine: engine}, nil
	case "step.approval_check":
		return &stepApprovalCheck{engine: engine}, nil
	case "step.approval_decide":
		return &stepApprovalDecide{engine: engine}, nil
	case "step.approval_list":
		return &stepApprovalList{engine: engine}, nil
	case "step.approval_escalate":
		return &stepApprovalEscalate{engine: engine}, nil
	case "step.approval_wait":
		return &stepApprovalWait{engine: engine}, nil
	default:
		return nil, fmt.Errorf("unknown step type: %s", typeName)
	}
}

func (p *plugin) CreateTypedStep(typeName, name string, config *anypb.Any) (sdk.StepInstance, error) {
	switch typeName {
	case "step.approval_request":
		cfg, err := typedStepConfig(config, &contracts.ApprovalCreateConfig{})
		if err != nil {
			return nil, err
		}
		engine := p.resolveEngineByName(cfg.GetEngine())
		if engine == nil {
			return nil, fmt.Errorf("no approval.engine module found; ensure one is defined in config")
		}
		factory := sdk.NewTypedStepFactory(typeName, &contracts.ApprovalCreateConfig{}, &contracts.ApprovalCreateArgs{}, typedApprovalCreate(engine))
		return factory.CreateTypedStep(typeName, name, config)
	case "step.approval_check":
		cfg, err := typedStepConfig(config, &contracts.ApprovalGetConfig{})
		if err != nil {
			return nil, err
		}
		engine := p.resolveEngineByName(cfg.GetEngine())
		if engine == nil {
			return nil, fmt.Errorf("no approval.engine module found; ensure one is defined in config")
		}
		factory := sdk.NewTypedStepFactory(typeName, &contracts.ApprovalGetConfig{}, &contracts.ApprovalGetArgs{}, typedApprovalGet(engine))
		return factory.CreateTypedStep(typeName, name, config)
	case "step.approval_decide":
		cfg, err := typedStepConfig(config, &contracts.ApprovalDecideConfig{})
		if err != nil {
			return nil, err
		}
		engine := p.resolveEngineByName(cfg.GetEngine())
		if engine == nil {
			return nil, fmt.Errorf("no approval.engine module found; ensure one is defined in config")
		}
		factory := sdk.NewTypedStepFactory(typeName, &contracts.ApprovalDecideConfig{}, &contracts.ApprovalDecideArgs{}, typedApprovalDecide(engine))
		return factory.CreateTypedStep(typeName, name, config)
	case "step.approval_list":
		cfg, err := typedStepConfig(config, &contracts.ApprovalListConfig{})
		if err != nil {
			return nil, err
		}
		engine := p.resolveEngineByName(cfg.GetEngine())
		if engine == nil {
			return nil, fmt.Errorf("no approval.engine module found; ensure one is defined in config")
		}
		factory := sdk.NewTypedStepFactory(typeName, &contracts.ApprovalListConfig{}, &contracts.ApprovalListArgs{}, typedApprovalList(engine))
		return factory.CreateTypedStep(typeName, name, config)
	case "step.approval_escalate":
		cfg, err := typedStepConfig(config, &contracts.ApprovalEscalateConfig{})
		if err != nil {
			return nil, err
		}
		engine := p.resolveEngineByName(cfg.GetEngine())
		if engine == nil {
			return nil, fmt.Errorf("no approval.engine module found; ensure one is defined in config")
		}
		factory := sdk.NewTypedStepFactory(typeName, &contracts.ApprovalEscalateConfig{}, &contracts.ApprovalEscalateArgs{}, typedApprovalEscalate(engine))
		return factory.CreateTypedStep(typeName, name, config)
	case "step.approval_wait":
		cfg, err := typedStepConfig(config, &contracts.ApprovalWaitConfig{})
		if err != nil {
			return nil, err
		}
		engine := p.resolveEngineByName(cfg.GetEngine())
		if engine == nil {
			return nil, fmt.Errorf("no approval.engine module found; ensure one is defined in config")
		}
		factory := sdk.NewTypedStepFactory(typeName, &contracts.ApprovalWaitConfig{}, &contracts.ApprovalWaitArgs{}, typedApprovalWait(engine))
		return factory.CreateTypedStep(typeName, name, config)
	default:
		return nil, fmt.Errorf("unknown step type: %s", typeName)
	}
}

func (p *plugin) ContractRegistry() *pb.ContractRegistry {
	const pkg = "workflow.plugins.approval.v1."
	contractsList := []*pb.ContractDescriptor{
		{
			Kind:          pb.ContractKind_CONTRACT_KIND_MODULE,
			ModuleType:    "approval.engine",
			ConfigMessage: pkg + "ApprovalEngineConfig",
			Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
		},
		stepContract("step.approval_request", "ApprovalCreateConfig", "ApprovalCreateArgs", "ApprovalRequest"),
		stepContract("step.approval_check", "ApprovalGetConfig", "ApprovalGetArgs", "ApprovalRequest"),
		stepContract("step.approval_decide", "ApprovalDecideConfig", "ApprovalDecideArgs", "ApprovalRequest"),
		stepContract("step.approval_list", "ApprovalListConfig", "ApprovalListArgs", "ApprovalListOutput"),
		stepContract("step.approval_escalate", "ApprovalEscalateConfig", "ApprovalEscalateArgs", "ApprovalRequest"),
		stepContract("step.approval_wait", "ApprovalWaitConfig", "ApprovalWaitArgs", "ApprovalRequest"),
		serviceContract("create", "ApprovalCreateArgs", "ApprovalRequest"),
		serviceContract("get", "ApprovalGetArgs", "ApprovalRequest"),
		serviceContract("decide", "ApprovalDecideArgs", "ApprovalRequest"),
		serviceContract("list", "ApprovalListArgs", "ApprovalListOutput"),
		serviceContract("escalate", "ApprovalEscalateArgs", "ApprovalRequest"),
		serviceContract("check_expiry", "ApprovalCheckExpiryArgs", "ApprovalCheckExpiryOutput"),
	}
	return &pb.ContractRegistry{
		Contracts: contractsList,
		FileDescriptorSet: &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{
			protodesc.ToFileDescriptorProto(contracts.File_internal_contracts_approval_proto),
		}},
	}
}

func stepContract(stepType, configMessage, inputMessage, outputMessage string) *pb.ContractDescriptor {
	const pkg = "workflow.plugins.approval.v1."
	return &pb.ContractDescriptor{
		Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
		StepType:      stepType,
		ConfigMessage: pkg + configMessage,
		InputMessage:  pkg + inputMessage,
		OutputMessage: pkg + outputMessage,
		Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
	}
}

func serviceContract(method, inputMessage, outputMessage string) *pb.ContractDescriptor {
	const pkg = "workflow.plugins.approval.v1."
	return &pb.ContractDescriptor{
		Kind:          pb.ContractKind_CONTRACT_KIND_SERVICE,
		ModuleType:    "approval.engine",
		ServiceName:   "approval.engine",
		Method:        method,
		InputMessage:  pkg + inputMessage,
		OutputMessage: pkg + outputMessage,
		Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
	}
}

func (p *plugin) resolveEngine(config map[string]any) *approvalEngine {
	if name, ok := config["engine"].(string); ok {
		return p.resolveEngineByName(name)
	}
	return p.resolveEngineByName("")
}

func (p *plugin) resolveEngineByName(name string) *approvalEngine {
	if name != "" {
		p.mu.RLock()
		defer p.mu.RUnlock()
		if e, found := p.engines[name]; found {
			return e
		}
		return nil
	}
	// Return the first (and typically only) engine.
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, e := range p.engines {
		return e
	}
	return nil
}

func typedStepConfig[T proto.Message](payload *anypb.Any, target T) (T, error) {
	if payload == nil {
		return target, nil
	}
	if payload.MessageName() != target.ProtoReflect().Descriptor().FullName() {
		var zero T
		return zero, fmt.Errorf("typed config type mismatch: expected %s, got %s", target.ProtoReflect().Descriptor().FullName(), payload.MessageName())
	}
	if err := payload.UnmarshalTo(target); err != nil {
		var zero T
		return zero, err
	}
	return target, nil
}
