package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GoCodeAlone/workflow-plugin-approval/internal/contracts"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func typedApprovalCreate(engine *approvalEngine) sdk.TypedStepHandler[*contracts.ApprovalCreateConfig, *contracts.ApprovalCreateArgs, *contracts.ApprovalRequest] {
	return func(_ context.Context, req sdk.TypedStepRequest[*contracts.ApprovalCreateConfig, *contracts.ApprovalCreateArgs]) (*sdk.TypedStepResult[*contracts.ApprovalRequest], error) {
		args, err := mergeProtoArgs(req.Config, req.Input)
		if err != nil {
			return nil, err
		}
		out, err := engine.InvokeMethod("create", args)
		if err != nil {
			return &sdk.TypedStepResult[*contracts.ApprovalRequest]{Output: &contracts.ApprovalRequest{Error: err.Error()}}, nil
		}
		return &sdk.TypedStepResult[*contracts.ApprovalRequest]{Output: approvalRequestFromMap(out)}, nil
	}
}

func typedApprovalGet(engine *approvalEngine) sdk.TypedStepHandler[*contracts.ApprovalGetConfig, *contracts.ApprovalGetArgs, *contracts.ApprovalRequest] {
	return func(_ context.Context, req sdk.TypedStepRequest[*contracts.ApprovalGetConfig, *contracts.ApprovalGetArgs]) (*sdk.TypedStepResult[*contracts.ApprovalRequest], error) {
		args, err := mergeProtoArgs(req.Config, req.Input)
		if err != nil {
			return nil, err
		}
		out, err := engine.InvokeMethod("get", args)
		if err != nil {
			return &sdk.TypedStepResult[*contracts.ApprovalRequest]{Output: &contracts.ApprovalRequest{Error: err.Error()}}, nil
		}
		return &sdk.TypedStepResult[*contracts.ApprovalRequest]{Output: approvalRequestFromMap(out)}, nil
	}
}

func typedApprovalDecide(engine *approvalEngine) sdk.TypedStepHandler[*contracts.ApprovalDecideConfig, *contracts.ApprovalDecideArgs, *contracts.ApprovalRequest] {
	return func(_ context.Context, req sdk.TypedStepRequest[*contracts.ApprovalDecideConfig, *contracts.ApprovalDecideArgs]) (*sdk.TypedStepResult[*contracts.ApprovalRequest], error) {
		args, err := mergeProtoArgs(req.Config, req.Input)
		if err != nil {
			return nil, err
		}
		out, err := engine.InvokeMethod("decide", args)
		if err != nil {
			return &sdk.TypedStepResult[*contracts.ApprovalRequest]{Output: &contracts.ApprovalRequest{Error: err.Error()}}, nil
		}
		return &sdk.TypedStepResult[*contracts.ApprovalRequest]{Output: approvalRequestFromMap(out)}, nil
	}
}

func typedApprovalList(engine *approvalEngine) sdk.TypedStepHandler[*contracts.ApprovalListConfig, *contracts.ApprovalListArgs, *contracts.ApprovalListOutput] {
	return func(_ context.Context, req sdk.TypedStepRequest[*contracts.ApprovalListConfig, *contracts.ApprovalListArgs]) (*sdk.TypedStepResult[*contracts.ApprovalListOutput], error) {
		args, err := mergeProtoArgs(req.Config, req.Input)
		if err != nil {
			return nil, err
		}
		out, err := engine.InvokeMethod("list", args)
		if err != nil {
			return &sdk.TypedStepResult[*contracts.ApprovalListOutput]{Output: &contracts.ApprovalListOutput{Error: err.Error()}}, nil
		}
		return &sdk.TypedStepResult[*contracts.ApprovalListOutput]{Output: approvalListFromMap(out)}, nil
	}
}

func typedApprovalEscalate(engine *approvalEngine) sdk.TypedStepHandler[*contracts.ApprovalEscalateConfig, *contracts.ApprovalEscalateArgs, *contracts.ApprovalRequest] {
	return func(_ context.Context, req sdk.TypedStepRequest[*contracts.ApprovalEscalateConfig, *contracts.ApprovalEscalateArgs]) (*sdk.TypedStepResult[*contracts.ApprovalRequest], error) {
		args, err := mergeProtoArgs(req.Config, req.Input)
		if err != nil {
			return nil, err
		}
		out, err := engine.InvokeMethod("escalate", args)
		if err != nil {
			return &sdk.TypedStepResult[*contracts.ApprovalRequest]{Output: &contracts.ApprovalRequest{Error: err.Error()}}, nil
		}
		return &sdk.TypedStepResult[*contracts.ApprovalRequest]{Output: approvalRequestFromMap(out)}, nil
	}
}

func typedApprovalWait(engine *approvalEngine) sdk.TypedStepHandler[*contracts.ApprovalWaitConfig, *contracts.ApprovalWaitArgs, *contracts.ApprovalRequest] {
	return func(ctx context.Context, req sdk.TypedStepRequest[*contracts.ApprovalWaitConfig, *contracts.ApprovalWaitArgs]) (*sdk.TypedStepResult[*contracts.ApprovalRequest], error) {
		step := &stepApprovalWait{engine: engine}
		args, err := mergeProtoArgs(req.Config, req.Input)
		if err != nil {
			return nil, err
		}
		result, err := step.Execute(ctx, nil, nil, args, nil, nil)
		if err != nil {
			return nil, err
		}
		out := approvalRequestFromMap(result.Output)
		return &sdk.TypedStepResult[*contracts.ApprovalRequest]{
			Output:       out,
			StopPipeline: result.StopPipeline,
		}, nil
	}
}

func (e *approvalEngine) InvokeTypedMethod(method string, input *anypb.Any) (*anypb.Any, error) {
	switch method {
	case "create":
		args, err := unpackTypedArgs(input, &contracts.ApprovalCreateArgs{})
		if err != nil {
			return nil, err
		}
		argMap, err := protoMessageToMap(args)
		if err != nil {
			return nil, err
		}
		out, err := e.InvokeMethod(method, argMap)
		if err != nil {
			return nil, err
		}
		return anypb.New(approvalRequestFromMap(out))
	case "get":
		args, err := unpackTypedArgs(input, &contracts.ApprovalGetArgs{})
		if err != nil {
			return nil, err
		}
		argMap, err := protoMessageToMap(args)
		if err != nil {
			return nil, err
		}
		out, err := e.InvokeMethod(method, argMap)
		if err != nil {
			return nil, err
		}
		return anypb.New(approvalRequestFromMap(out))
	case "decide":
		args, err := unpackTypedArgs(input, &contracts.ApprovalDecideArgs{})
		if err != nil {
			return nil, err
		}
		argMap, err := protoMessageToMap(args)
		if err != nil {
			return nil, err
		}
		out, err := e.InvokeMethod(method, argMap)
		if err != nil {
			return nil, err
		}
		return anypb.New(approvalRequestFromMap(out))
	case "list":
		args, err := unpackTypedArgs(input, &contracts.ApprovalListArgs{})
		if err != nil {
			return nil, err
		}
		argMap, err := protoMessageToMap(args)
		if err != nil {
			return nil, err
		}
		out, err := e.InvokeMethod(method, argMap)
		if err != nil {
			return nil, err
		}
		return anypb.New(approvalListFromMap(out))
	case "escalate":
		args, err := unpackTypedArgs(input, &contracts.ApprovalEscalateArgs{})
		if err != nil {
			return nil, err
		}
		argMap, err := protoMessageToMap(args)
		if err != nil {
			return nil, err
		}
		out, err := e.InvokeMethod(method, argMap)
		if err != nil {
			return nil, err
		}
		return anypb.New(approvalRequestFromMap(out))
	case "check_expiry":
		if _, err := unpackTypedArgs(input, &contracts.ApprovalCheckExpiryArgs{}); err != nil {
			return nil, err
		}
		out, err := e.InvokeMethod(method, nil)
		if err != nil {
			return nil, err
		}
		return anypb.New(approvalExpiryFromMap(out))
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func unpackTypedArgs[T proto.Message](input *anypb.Any, target T) (T, error) {
	if input == nil {
		var zero T
		return zero, fmt.Errorf("typed input is required")
	}
	if input.MessageName() != target.ProtoReflect().Descriptor().FullName() {
		var zero T
		return zero, fmt.Errorf("typed input type mismatch: expected %s, got %s", target.ProtoReflect().Descriptor().FullName(), input.MessageName())
	}
	if err := input.UnmarshalTo(target); err != nil {
		var zero T
		return zero, err
	}
	return target, nil
}

func mergeProtoArgs(config, input proto.Message) (map[string]any, error) {
	args, err := protoMessageToMap(config)
	if err != nil {
		return nil, err
	}
	if args == nil {
		args = make(map[string]any)
	}
	inputArgs, err := protoMessageToMap(input)
	if err != nil {
		return nil, err
	}
	for k, v := range inputArgs {
		args[k] = v
	}
	return args, nil
}

func protoMessageToMap(msg proto.Message) (map[string]any, error) {
	if msg == nil {
		return nil, nil
	}
	raw, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal typed protobuf args: %w", err)
	}
	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("decode typed protobuf args: %w", err)
	}
	return values, nil
}

func approvalRequestFromMap(values map[string]any) *contracts.ApprovalRequest {
	if values == nil {
		return &contracts.ApprovalRequest{}
	}
	out := &contracts.ApprovalRequest{
		Id:                firstStringVal(values, "id", "request_id"),
		PipelineId:        strVal(values, "pipeline_id"),
		Title:             strVal(values, "title"),
		Description:       strVal(values, "description"),
		RequiredApprovals: int32Val(values, "required_approvals"),
		Status:            strVal(values, "status"),
		ContinuationToken: strVal(values, "continuation_token"),
		CreatedAt:         strVal(values, "created_at"),
		ExpiresAt:         strVal(values, "expires_at"),
		Error:             strVal(values, "error"),
		TimedOut:          boolVal(values, "timed_out"),
	}
	out.Approvers = stringSliceVal(values["approvers"])
	for _, decision := range mapSliceVal(values["decisions"]) {
		out.Decisions = append(out.Decisions, &contracts.ApprovalDecision{
			Actor:     strVal(decision, "actor"),
			Decision:  strVal(decision, "decision"),
			Comment:   strVal(decision, "comment"),
			Timestamp: strVal(decision, "timestamp"),
		})
	}
	return out
}

func approvalListFromMap(values map[string]any) *contracts.ApprovalListOutput {
	out := &contracts.ApprovalListOutput{
		Count: int32Val(values, "count"),
		Error: strVal(values, "error"),
	}
	for _, request := range mapSliceVal(values["requests"]) {
		out.Requests = append(out.Requests, approvalRequestFromMap(request))
	}
	return out
}

func approvalExpiryFromMap(values map[string]any) *contracts.ApprovalCheckExpiryOutput {
	return &contracts.ApprovalCheckExpiryOutput{
		Expired: stringSliceVal(values["expired"]),
		Count:   int32Val(values, "count"),
		Error:   strVal(values, "error"),
	}
}

func int32Val(values map[string]any, key string) int32 {
	if n, ok := toInt(values[key]); ok {
		return int32(n)
	}
	return 0
}

func boolVal(values map[string]any, key string) bool {
	v, _ := values[key].(bool)
	return v
}

func firstStringVal(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strVal(values, key); value != "" {
			return value
		}
	}
	return ""
}

func stringSliceVal(value any) []string {
	switch items := value.(type) {
	case []string:
		return append([]string(nil), items...)
	case []any:
		out := make([]string, 0, len(items))
		for _, item := range items {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func mapSliceVal(value any) []map[string]any {
	switch items := value.(type) {
	case []map[string]any:
		return items
	case []any:
		out := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}
