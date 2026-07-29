package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// Status represents a state in the state machine
type Status string

// Transition defines a valid move from one state to another
type Transition struct {
	From   Status
	To     Status
	Guard  GuardFunc
	Effect EffectFunc
	Events []string
}

// GuardFunc validates whether a transition is allowed
type GuardFunc func(ctx context.Context, resourceID uuid.UUID) error

// EffectFunc performs side effects after a successful transition
type EffectFunc func(ctx context.Context, resourceID uuid.UUID) error

// GuardError wraps multiple guard failures
type GuardError struct {
	Reasons []string
}

func (e *GuardError) Error() string {
	return fmt.Sprintf("guard check failed: %v", e.Reasons)
}

// InvalidTransitionError is returned when a transition is not defined
type InvalidTransitionError struct {
	ResourceType string
	ResourceID   uuid.UUID
	From, To     Status
}

func (e *InvalidTransitionError) Error() string {
	return fmt.Sprintf("invalid transition for %s %s: %s → %s",
		e.ResourceType, e.ResourceID, e.From, e.To)
}

// StateMachine manages transitions for a resource type
type StateMachine struct {
	mu           sync.RWMutex
	resourceType string
	db           *pgxpool.Pool
	eventBus     *nats.Conn
	transitions  map[Status]map[Status]*Transition
}

// NewStateMachine creates a new state machine
func NewStateMachine(resourceType string, db *pgxpool.Pool, nc *nats.Conn) *StateMachine {
	return &StateMachine{
		resourceType: resourceType,
		db:           db,
		eventBus:     nc,
		transitions:  make(map[Status]map[Status]*Transition),
	}
}

// AddTransition registers a new transition
func (sm *StateMachine) AddTransition(t Transition) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.transitions[t.From] == nil {
		sm.transitions[t.From] = make(map[Status]*Transition)
	}
	sm.transitions[t.From][t.To] = &t
}

// GetTransition returns the transition definition if it exists
func (sm *StateMachine) GetTransition(from, to Status) (*Transition, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if fromMap, ok := sm.transitions[from]; ok {
		if t, ok := fromMap[to]; ok {
			return t, true
		}
	}
	return nil, false
}

// CanTransition checks if a transition is possible (guard only)
func (sm *StateMachine) CanTransition(ctx context.Context, resourceID uuid.UUID, to Status) (bool, error) {
	current, err := sm.getCurrentState(ctx, resourceID)
	if err != nil {
		return false, err
	}

	t, ok := sm.GetTransition(current, to)
	if !ok {
		return false, &InvalidTransitionError{
			ResourceType: sm.resourceType,
			ResourceID:   resourceID,
			From:         current,
			To:           to,
		}
	}

	if t.Guard != nil {
		if err := t.Guard(ctx, resourceID); err != nil {
			return false, err
		}
	}

	return true, nil
}

// Execute performs a state transition with guard check, state update, and side effects
func (sm *StateMachine) Execute(ctx context.Context, resourceID uuid.UUID, to Status) error {
	current, err := sm.getCurrentState(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	t, ok := sm.GetTransition(current, to)
	if !ok {
		return &InvalidTransitionError{
			ResourceType: sm.resourceType,
			ResourceID:   resourceID,
			From:         current,
			To:           to,
		}
	}

	// 1. Guard check
	if t.Guard != nil {
		if err := t.Guard(ctx, resourceID); err != nil {
			slog.Warn("guard check failed",
				"resource_type", sm.resourceType,
				"resource_id", resourceID,
				"from", current,
				"to", to,
				"error", err,
			)
			return fmt.Errorf("guard check failed: %w", err)
		}
	}

	// 2. Apply state change
	if err := sm.applyState(ctx, resourceID, current, to); err != nil {
		return fmt.Errorf("failed to apply state change: %w", err)
	}

	slog.Info("state transition applied",
		"resource_type", sm.resourceType,
		"resource_id", resourceID,
		"from", current,
		"to", to,
	)

	// 3. Side effects — fire and forget (non-blocking)
	if t.Effect != nil {
		go func() {
			bgCtx := context.Background()
			if err := t.Effect(bgCtx, resourceID); err != nil {
				slog.Error("side effect failed",
					"resource_type", sm.resourceType,
					"resource_id", resourceID,
					"transition", fmt.Sprintf("%s→%s", current, to),
					"error", err,
				)
			}
		}()
	}

	// 4. Publish events to NATS
	event := StateChangeEvent{
		ResourceType: sm.resourceType,
		ResourceID:   resourceID,
		From:         string(current),
		To:           string(to),
		Timestamp:    time.Now(),
	}

	for _, subject := range t.Events {
		pubSubject := fmt.Sprintf("%s.%s", subject, resourceID.String())
		go func(s string) {
			data, _ := jsonMarshal(event)
			if err := sm.eventBus.Publish(s, data); err != nil {
				slog.Error("failed to publish event",
					"subject", s,
					"error", err,
				)
			}
		}(pubSubject)
	}

	return nil
}

// StateChangeEvent is published after each transition
type StateChangeEvent struct {
	ResourceType string    `json:"resource_type"`
	ResourceID   uuid.UUID `json:"resource_id"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	Timestamp    time.Time `json:"timestamp"`
}

// getCurrentState retrieves the current state from the appropriate table
func (sm *StateMachine) getCurrentState(ctx context.Context, resourceID uuid.UUID) (Status, error) {
	var status string
	var query string

	switch sm.resourceType {
	case "case":
		query = `SELECT status::text FROM cases WHERE id = $1`
	case "map_service":
		query = `SELECT status::text FROM map_services WHERE id = $1`
	case "claim_service":
		query = `SELECT status::text FROM claim_services WHERE id = $1`
	case "cert_service":
		query = `SELECT status::text FROM cert_services WHERE id = $1`
	default:
		return "", fmt.Errorf("unknown resource type: %s", sm.resourceType)
	}

	err := sm.db.QueryRow(ctx, query, resourceID).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("resource not found: %w", err)
	}

	return Status(status), nil
}

// applyState updates the resource status in the database
func (sm *StateMachine) applyState(ctx context.Context, resourceID uuid.UUID, from, to Status) error {
	var query string

	switch sm.resourceType {
	case "case":
		query = `UPDATE cases SET status = $1::case_status, updated_at = NOW() WHERE id = $2 AND status::text = $3`
	case "map_service":
		query = `UPDATE map_services SET status = $1::map_service_status, updated_at = NOW() WHERE id = $2 AND status::text = $3`
	case "claim_service":
		query = `UPDATE claim_services SET status = $1::claim_service_status, updated_at = NOW() WHERE id = $2 AND status::text = $3`
	case "cert_service":
		query = `UPDATE cert_services SET status = $1::cert_service_status, updated_at = NOW() WHERE id = $2 AND status::text = $3`
	default:
		return fmt.Errorf("unknown resource type: %s", sm.resourceType)
	}

	tag, err := sm.db.Exec(ctx, query, string(to), resourceID, string(from))
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("state change was not applied — possible race condition")
	}

	return nil
}

// JSON marshal helper
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
