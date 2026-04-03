package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// approvalEngine implements ModuleInstance, ServiceInvoker, and MessageAwareModule.
type approvalEngine struct {
	store      *Store
	publisher  sdk.MessagePublisher
	subscriber sdk.MessageSubscriber
	config     map[string]any
}

// NewApprovalEngine creates a new approval engine module instance.
func NewApprovalEngine(config map[string]any) *approvalEngine {
	return &approvalEngine{config: config}
}

func (e *approvalEngine) Init() error {
	e.store = NewStore()
	return nil
}

func (e *approvalEngine) Start(_ context.Context) error { return nil }

func (e *approvalEngine) Stop(_ context.Context) error { return nil }

// SetMessagePublisher receives the host's message publisher.
func (e *approvalEngine) SetMessagePublisher(pub sdk.MessagePublisher) {
	e.publisher = pub
}

// SetMessageSubscriber receives the host's message subscriber.
func (e *approvalEngine) SetMessageSubscriber(sub sdk.MessageSubscriber) {
	e.subscriber = sub
}

func (e *approvalEngine) publish(topic string, data any) {
	if e.publisher == nil {
		return
	}
	payload, _ := json.Marshal(data)
	_, _ = e.publisher.Publish(topic, payload, map[string]string{"source": "approval.engine"})
}

// InvokeMethod dispatches service calls to the engine.
func (e *approvalEngine) InvokeMethod(method string, args map[string]any) (map[string]any, error) {
	switch method {
	case "create":
		return e.invokeCreate(args)
	case "get":
		return e.invokeGet(args)
	case "decide":
		return e.invokeDecide(args)
	case "list":
		return e.invokeList(args)
	case "escalate":
		return e.invokeEscalate(args)
	case "check_expiry":
		return e.invokeCheckExpiry()
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func (e *approvalEngine) invokeCreate(args map[string]any) (map[string]any, error) {
	req := &ApprovalRequest{
		ID:          strVal(args, "id"),
		PipelineID:  strVal(args, "pipeline_id"),
		Title:       strVal(args, "title"),
		Description: strVal(args, "description"),
	}
	if approvers, ok := args["approvers"].([]any); ok {
		for _, a := range approvers {
			if s, ok := a.(string); ok {
				req.Approvers = append(req.Approvers, s)
			}
		}
	}
	if n, ok := toInt(args["required_approvals"]); ok {
		req.RequiredApprovals = n
	}
	if ttl, ok := args["ttl_seconds"].(float64); ok && ttl > 0 {
		req.ExpiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
	}
	created, err := e.store.Create(req)
	if err != nil {
		return nil, err
	}
	e.publish("approval.requested", created)
	return requestToMap(created), nil
}

func (e *approvalEngine) invokeGet(args map[string]any) (map[string]any, error) {
	id := strVal(args, "request_id")
	if id == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	req, err := e.store.Get(id)
	if err != nil {
		return nil, err
	}
	return requestToMap(req), nil
}

func (e *approvalEngine) invokeDecide(args map[string]any) (map[string]any, error) {
	id := strVal(args, "request_id")
	if id == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	decision := ApprovalDecision{
		Actor:    strVal(args, "actor"),
		Decision: strVal(args, "decision"),
		Comment:  strVal(args, "comment"),
	}
	if decision.Actor == "" || decision.Decision == "" {
		return nil, fmt.Errorf("actor and decision are required")
	}
	if decision.Decision != "approve" && decision.Decision != "reject" {
		return nil, fmt.Errorf("decision must be 'approve' or 'reject'")
	}
	updated, err := e.store.UpdateDecision(id, decision)
	if err != nil {
		return nil, err
	}
	e.publish("approval.decided", map[string]any{
		"request_id": id,
		"actor":      decision.Actor,
		"decision":   decision.Decision,
		"status":     string(updated.Status),
	})
	return requestToMap(updated), nil
}

func (e *approvalEngine) invokeList(args map[string]any) (map[string]any, error) {
	approver := strVal(args, "approver")
	requests := e.store.List(approver)
	items := make([]map[string]any, 0, len(requests))
	for _, r := range requests {
		items = append(items, requestToMap(r))
	}
	return map[string]any{"requests": items, "count": len(items)}, nil
}

func (e *approvalEngine) invokeEscalate(args map[string]any) (map[string]any, error) {
	id := strVal(args, "request_id")
	if id == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	var newApprovers []string
	if approvers, ok := args["new_approvers"].([]any); ok {
		for _, a := range approvers {
			if s, ok := a.(string); ok {
				newApprovers = append(newApprovers, s)
			}
		}
	}
	updated, err := e.store.Escalate(id, newApprovers)
	if err != nil {
		return nil, err
	}
	e.publish("approval.escalated", map[string]any{
		"request_id":    id,
		"new_approvers": newApprovers,
	})
	return requestToMap(updated), nil
}

func (e *approvalEngine) invokeCheckExpiry() (map[string]any, error) {
	expired := e.store.CheckExpiry(time.Now())
	for _, id := range expired {
		e.publish("approval.expired", map[string]any{"request_id": id})
	}
	return map[string]any{"expired": expired, "count": len(expired)}, nil
}

// GetStore exposes the store for step implementations.
func (e *approvalEngine) GetStore() *Store { return e.store }

func requestToMap(r *ApprovalRequest) map[string]any {
	decisions := make([]map[string]any, 0, len(r.Decisions))
	for _, d := range r.Decisions {
		decisions = append(decisions, map[string]any{
			"actor":     d.Actor,
			"decision":  d.Decision,
			"comment":   d.Comment,
			"timestamp": d.Timestamp.Format(time.RFC3339),
		})
	}
	m := map[string]any{
		"id":                 r.ID,
		"pipeline_id":       r.PipelineID,
		"title":             r.Title,
		"description":       r.Description,
		"approvers":         r.Approvers,
		"required_approvals": r.RequiredApprovals,
		"status":            string(r.Status),
		"decisions":         decisions,
		"continuation_token": r.ContinuationToken,
		"created_at":        r.CreatedAt.Format(time.RFC3339),
	}
	if !r.ExpiresAt.IsZero() {
		m["expires_at"] = r.ExpiresAt.Format(time.RFC3339)
	}
	return m
}

func strVal(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}
