package internal

import (
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// Manifest returns the plugin metadata used by the workflow engine for
// discovery and capability negotiation.
var Manifest = sdk.PluginManifest{
	Name:        "workflow-plugin-approval",
	Version:     "0.1.0",
	Description: "Human-in-the-loop approval workflows for the workflow engine",
	Author:      "GoCodeAlone",
}

// plugin implements PluginProvider, ModuleProvider, and StepProvider.
type plugin struct {
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

func (p *plugin) CreateModule(typeName, name string, config map[string]any) (sdk.ModuleInstance, error) {
	if typeName != "approval.engine" {
		return nil, fmt.Errorf("unknown module type: %s", typeName)
	}
	engine := NewApprovalEngine(config)
	p.engines[name] = engine
	return engine, nil
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

func (p *plugin) resolveEngine(config map[string]any) *approvalEngine {
	if name, ok := config["engine"].(string); ok {
		if e, found := p.engines[name]; found {
			return e
		}
	}
	// Return the first (and typically only) engine.
	for _, e := range p.engines {
		return e
	}
	return nil
}
