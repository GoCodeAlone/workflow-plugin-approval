# CLAUDE.md — Workflow Plugin Approval

Human-in-the-loop approval workflows for the GoCodeAlone/workflow engine.
In-memory store, state machine (pending→approved|rejected|escalated|expired),
EventBus integration, and ServiceInvoker support.

## Build & Test

```sh
go build ./...
go test ./... -v -race -count=1
```

## Cross-compile for deployment

```sh
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o workflow-plugin-approval ./cmd/workflow-plugin-approval/
```

## Structure

- `cmd/workflow-plugin-approval/main.go` — Plugin entry point (calls `sdk.Serve`)
- `internal/plugin.go` — Plugin manifest, module/step providers
- `internal/approval.go` — Core types (ApprovalRequest, ApprovalDecision, ApprovalStatus)
- `internal/store.go` — Thread-safe in-memory approval store
- `internal/module_engine.go` — approval.engine module (ServiceInvoker + MessageAwareModule)
- `internal/steps.go` — All 6 step type implementations
- `plugin.json` — Capability manifest for the workflow registry

## Module: approval.engine

ServiceInvoker methods: create, get, decide, list, escalate, check_expiry
EventBus topics: approval.requested, approval.decided, approval.escalated, approval.expired

## Steps

- `step.approval_request` — Create a new approval request
- `step.approval_check` — Check status of an approval request
- `step.approval_decide` — Record approve/reject decision
- `step.approval_list` — List approvals (filterable by approver)
- `step.approval_escalate` — Reassign approvers
- `step.approval_wait` — Poll until decided or expired

## Releasing

```sh
git tag v0.1.0
git push origin v0.1.0
```
