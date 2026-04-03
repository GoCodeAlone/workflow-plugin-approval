package internal

import "time"

// ApprovalStatus represents the state of an approval request.
type ApprovalStatus string

const (
	StatusPending   ApprovalStatus = "pending"
	StatusApproved  ApprovalStatus = "approved"
	StatusRejected  ApprovalStatus = "rejected"
	StatusEscalated ApprovalStatus = "escalated"
	StatusExpired   ApprovalStatus = "expired"
)

// ApprovalRequest is a single approval request tracked by the engine.
type ApprovalRequest struct {
	ID                string             `json:"id"`
	PipelineID        string             `json:"pipeline_id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	Approvers         []string           `json:"approvers"`
	RequiredApprovals int                `json:"required_approvals"`
	Status            ApprovalStatus     `json:"status"`
	Decisions         []ApprovalDecision `json:"decisions"`
	ContinuationToken string             `json:"continuation_token"`
	CreatedAt         time.Time          `json:"created_at"`
	ExpiresAt         time.Time          `json:"expires_at"`
}

// ApprovalDecision records a single actor's approve/reject decision.
type ApprovalDecision struct {
	Actor     string    `json:"actor"`
	Decision  string    `json:"decision"` // "approve" or "reject"
	Comment   string    `json:"comment"`
	Timestamp time.Time `json:"timestamp"`
}

// IsTerminal returns true if the status is a final state.
func (s ApprovalStatus) IsTerminal() bool {
	return s == StatusApproved || s == StatusRejected || s == StatusExpired
}
