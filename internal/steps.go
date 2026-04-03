package internal

import (
	"context"
	"time"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// stepApprovalRequest creates a new approval request.
type stepApprovalRequest struct {
	engine *approvalEngine
}

func (s *stepApprovalRequest) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	args := mergeArgs(config, current)
	result, err := s.engine.InvokeMethod("create", args)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}
	return &sdk.StepResult{Output: result}, nil
}

// stepApprovalCheck checks the status of an existing approval request.
type stepApprovalCheck struct {
	engine *approvalEngine
}

func (s *stepApprovalCheck) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	args := mergeArgs(config, current)
	result, err := s.engine.InvokeMethod("get", args)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}
	return &sdk.StepResult{Output: result}, nil
}

// stepApprovalDecide records an approve/reject decision.
type stepApprovalDecide struct {
	engine *approvalEngine
}

func (s *stepApprovalDecide) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	args := mergeArgs(config, current)
	result, err := s.engine.InvokeMethod("decide", args)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}
	return &sdk.StepResult{Output: result}, nil
}

// stepApprovalList lists pending approvals.
type stepApprovalList struct {
	engine *approvalEngine
}

func (s *stepApprovalList) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	args := mergeArgs(config, current)
	result, err := s.engine.InvokeMethod("list", args)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}
	return &sdk.StepResult{Output: result}, nil
}

// stepApprovalEscalate reassigns approvers on a request.
type stepApprovalEscalate struct {
	engine *approvalEngine
}

func (s *stepApprovalEscalate) Execute(_ context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	args := mergeArgs(config, current)
	result, err := s.engine.InvokeMethod("escalate", args)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}
	return &sdk.StepResult{Output: result}, nil
}

// stepApprovalWait polls for approval completion.
type stepApprovalWait struct {
	engine *approvalEngine
}

func (s *stepApprovalWait) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	args := mergeArgs(config, current)
	reqID := strVal(args, "request_id")
	if reqID == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "request_id is required"}}, nil
	}

	pollMs := 500
	if p, ok := toInt(args["poll_interval_ms"]); ok && p > 0 {
		pollMs = p
	}
	timeoutSec := 300
	if t, ok := toInt(args["timeout_seconds"]); ok && t > 0 {
		timeoutSec = t
	}

	deadline := time.After(time.Duration(timeoutSec) * time.Second)
	ticker := time.NewTicker(time.Duration(pollMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		req, err := s.engine.store.Get(reqID)
		if err != nil {
			return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
		}
		if req.Status.IsTerminal() {
			return &sdk.StepResult{Output: requestToMap(req)}, nil
		}
		// Check for expiry
		if !req.ExpiresAt.IsZero() && time.Now().After(req.ExpiresAt) {
			s.engine.store.CheckExpiry(time.Now())
			req, _ = s.engine.store.Get(reqID)
			return &sdk.StepResult{Output: requestToMap(req)}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return &sdk.StepResult{
				Output: map[string]any{
					"request_id": reqID,
					"status":     string(req.Status),
					"timed_out":  true,
				},
				StopPipeline: true,
			}, nil
		case <-ticker.C:
			// poll again
		}
	}
}

func mergeArgs(config, current map[string]any) map[string]any {
	args := make(map[string]any)
	for k, v := range config {
		args[k] = v
	}
	for k, v := range current {
		args[k] = v
	}
	return args
}
