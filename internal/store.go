package internal

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Store is a thread-safe in-memory store for approval requests.
type Store struct {
	mu       sync.RWMutex
	requests map[string]*ApprovalRequest
}

// NewStore creates a new empty approval store.
func NewStore() *Store {
	return &Store{requests: make(map[string]*ApprovalRequest)}
}

// Create adds a new approval request and returns it.
func (s *Store) Create(req *ApprovalRequest) (*ApprovalRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.ID == "" {
		req.ID = generateID()
	}
	if _, exists := s.requests[req.ID]; exists {
		return nil, fmt.Errorf("approval request %s already exists", req.ID)
	}
	if req.RequiredApprovals <= 0 {
		req.RequiredApprovals = 1
	}
	if req.Status == "" {
		req.Status = StatusPending
	}
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}
	if req.ContinuationToken == "" {
		req.ContinuationToken = generateID()
	}
	s.requests[req.ID] = req
	return req, nil
}

// Get retrieves a request by ID. Returns a deep copy to prevent data races.
func (s *Store) Get(id string) (*ApprovalRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.requests[id]
	if !ok {
		return nil, fmt.Errorf("approval request %s not found", id)
	}
	return r.clone(), nil
}

// List returns all requests, optionally filtered by approver. Returns deep copies to prevent data races.
func (s *Store) List(approver string) []*ApprovalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*ApprovalRequest
	for _, r := range s.requests {
		if approver == "" {
			result = append(result, r.clone())
			continue
		}
		for _, a := range r.Approvers {
			if a == approver {
				result = append(result, r.clone())
				break
			}
		}
	}
	return result
}

// UpdateDecision records an approve/reject decision from an actor.
func (s *Store) UpdateDecision(id string, decision ApprovalDecision) (*ApprovalRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.requests[id]
	if !ok {
		return nil, fmt.Errorf("approval request %s not found", id)
	}
	if r.Status.IsTerminal() {
		return nil, fmt.Errorf("approval request %s is already %s", id, r.Status)
	}
	if !isAuthorizedApprover(r, decision.Actor) {
		return nil, fmt.Errorf("actor %s is not an authorized approver for request %s", decision.Actor, id)
	}
	for _, d := range r.Decisions {
		if d.Actor == decision.Actor {
			return nil, fmt.Errorf("actor %s has already decided on request %s", decision.Actor, id)
		}
	}
	if decision.Timestamp.IsZero() {
		decision.Timestamp = time.Now()
	}
	r.Decisions = append(r.Decisions, decision)
	r.Status = resolveStatus(r)
	return r, nil
}

// Escalate replaces the approvers list on a pending/escalated request.
func (s *Store) Escalate(id string, newApprovers []string) (*ApprovalRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.requests[id]
	if !ok {
		return nil, fmt.Errorf("approval request %s not found", id)
	}
	if r.Status.IsTerminal() {
		return nil, fmt.Errorf("cannot escalate request %s in %s state", id, r.Status)
	}
	if len(newApprovers) == 0 {
		return nil, fmt.Errorf("new approvers list cannot be empty")
	}
	r.Approvers = newApprovers
	r.Status = StatusEscalated
	return r, nil
}

// CheckExpiry marks pending/escalated requests as expired if past ExpiresAt.
// Returns the list of newly expired request IDs.
func (s *Store) CheckExpiry(now time.Time) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var expired []string
	for _, r := range s.requests {
		if !r.Status.IsTerminal() && r.Status != StatusExpired &&
			!r.ExpiresAt.IsZero() && now.After(r.ExpiresAt) {
			r.Status = StatusExpired
			expired = append(expired, r.ID)
		}
	}
	return expired
}

func resolveStatus(r *ApprovalRequest) ApprovalStatus {
	approvals, rejections, relevant := 0, 0, 0
	for _, d := range r.Decisions {
		if !contains(r.Approvers, d.Actor) {
			continue // skip stale decisions from removed approvers
		}
		relevant++
		switch d.Decision {
		case "approve":
			approvals++
		case "reject":
			rejections++
		}
	}
	if approvals >= r.RequiredApprovals {
		return StatusApproved
	}
	remaining := len(r.Approvers) - relevant
	if remaining+approvals < r.RequiredApprovals {
		return StatusRejected
	}
	if rejections > 0 && remaining == 0 {
		return StatusRejected
	}
	if r.Status == StatusEscalated {
		return StatusEscalated
	}
	return StatusPending
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func isAuthorizedApprover(r *ApprovalRequest, actor string) bool {
	for _, a := range r.Approvers {
		if a == actor {
			return true
		}
	}
	return false
}

func (r *ApprovalRequest) clone() *ApprovalRequest {
	cp := *r
	cp.Approvers = make([]string, len(r.Approvers))
	copy(cp.Approvers, r.Approvers)
	cp.Decisions = make([]ApprovalDecision, len(r.Decisions))
	copy(cp.Decisions, r.Decisions)
	return &cp
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
