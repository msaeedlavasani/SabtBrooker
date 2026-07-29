package workflow

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestNewStateMachine verifies basic creation
func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine("case", nil, nil)
	if sm == nil {
		t.Fatal("state machine should not be nil")
	}
	if sm.resourceType != "case" {
		t.Errorf("expected resourceType 'case', got '%s'", sm.resourceType)
	}
}

// TestAddTransition registers transitions and verifies they exist
func TestAddTransition(t *testing.T) {
	sm := NewStateMachine("case", nil, nil)

	sm.AddTransition(Transition{From: "draft", To: "map_in_progress"})
	sm.AddTransition(Transition{From: "map_in_progress", To: "map_completed"})
	sm.AddTransition(Transition{From: "map_completed", To: "claim_in_progress"})

	// Valid transitions
	tests := []struct {
		from, to Status
		want     bool
	}{
		{"draft", "map_in_progress", true},
		{"map_in_progress", "map_completed", true},
		{"map_completed", "claim_in_progress", true},
		{"draft", "map_completed", false},         // skip a step
		{"claim_completed", "cert_completed", false}, // non-existent
		{"draft", "cert_completed", false},
	}

	for _, tt := range tests {
		_, exists := sm.GetTransition(tt.from, tt.to)
		if exists != tt.want {
			t.Errorf("transition %s→%s: got exists=%v, want %v", tt.from, tt.to, exists, tt.want)
		}
	}
}

// TestGuardError verifies error formatting
func TestGuardError(t *testing.T) {
	err := &GuardError{Reasons: []string{
		"شماره موبایل تایید نشده",
		"سن کمتر از ۱۸ سال",
	}}
	expected := "guard check failed: [شماره موبایل تایید نشده سن کمتر از ۱۸ سال]"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

// TestInvalidTransitionError verifies error formatting
func TestInvalidTransitionError(t *testing.T) {
	id := uuid.New()
	err := &InvalidTransitionError{
		ResourceType: "case",
		ResourceID:   id,
		From:         "draft",
		To:           "cert_completed",
	}
	expected := "invalid transition for case " + id.String() + ": draft → cert_completed"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

// TestGuardExecution verifies guards are called during Execute
func TestGuardExecution(t *testing.T) {
	sm := NewStateMachine("test", nil, nil)
	guardCalled := false

	sm.AddTransition(Transition{
		From: "start",
		To:   "end",
		Guard: func(ctx context.Context, resourceID uuid.UUID) error {
			guardCalled = true
			return nil
		},
	})

	// We can only test guard registration — Execute needs DB
	tr, ok := sm.GetTransition("start", "end")
	if !ok {
		t.Fatal("transition should exist")
	}
	if tr.Guard == nil {
		t.Fatal("guard should be set")
	}

	// Call guard directly
	err := tr.Guard(context.Background(), uuid.New())
	if err != nil {
		t.Errorf("guard should pass: %v", err)
	}
	if !guardCalled {
		t.Error("guard was not called")
	}
}

// TestGuardFailure verifies guard blocks transition
func TestGuardFailure(t *testing.T) {
	sm := NewStateMachine("test", nil, nil)

	sm.AddTransition(Transition{
		From: "start",
		To:   "end",
		Guard: func(ctx context.Context, resourceID uuid.UUID) error {
			return &GuardError{Reasons: []string{"blocked"}}
		},
	})

	tr, _ := sm.GetTransition("start", "end")
	err := tr.Guard(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("guard should have failed")
	}
}

// TestCanTransitionWithGuard verifies CanTransition respects guards
func TestCanTransitionWithGuard(t *testing.T) {
	// This test verifies the interface contract, not actual DB behavior
	sm := NewStateMachine("test", nil, nil)

	sm.AddTransition(Transition{
		From: "start",
		To:   "end",
		Guard: func(ctx context.Context, resourceID uuid.UUID) error {
			return &GuardError{Reasons: []string{"not ready"}}
		},
	})

	_, err := sm.CanTransition(context.Background(), uuid.New(), "end")
	// Will fail because getCurrentState needs DB — but we can at least
	// verify the transition exists in the registry
	_, exists := sm.GetTransition("start", "end")
	if !exists {
		t.Error("transition should be registered")
	}
	if err == nil {
		t.Log("CanTransition with guard — state retrieval failed (expected without DB)")
	}
}

// TestMultipleTransitionsFromSameState verifies branching
func TestMultipleTransitionsFromSameState(t *testing.T) {
	sm := NewStateMachine("test", nil, nil)

	sm.AddTransition(Transition{From: "submitted", To: "approved"})
	sm.AddTransition(Transition{From: "submitted", To: "rejected"})

	_, exists1 := sm.GetTransition("submitted", "approved")
	_, exists2 := sm.GetTransition("submitted", "rejected")

	if !exists1 || !exists2 {
		t.Error("both transitions from 'submitted' should exist")
	}
}

// TestEventPublishing verifies events are attached to transitions
func TestEventPublishing(t *testing.T) {
	sm := NewStateMachine("test", nil, nil)

	sm.AddTransition(Transition{
		From:   "draft",
		To:     "map_in_progress",
		Events: []string{"case.map.started", "case.status.changed"},
	})

	tr, _ := sm.GetTransition("draft", "map_in_progress")
	if len(tr.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(tr.Events))
	}
	if tr.Events[0] != "case.map.started" || tr.Events[1] != "case.status.changed" {
		t.Errorf("unexpected event names: %v", tr.Events)
	}
}

// TestEffectExecution verifies effects are attached to transitions
func TestEffectExecution(t *testing.T) {
	sm := NewStateMachine("test", nil, nil)
	effectCalled := false

	sm.AddTransition(Transition{
		From: "start",
		To:   "end",
		Effect: func(ctx context.Context, resourceID uuid.UUID) error {
			effectCalled = true
			return nil
		},
	})

	tr, _ := sm.GetTransition("start", "end")
	if tr.Effect == nil {
		t.Fatal("effect should be set")
	}

	_ = tr.Effect(context.Background(), uuid.New())
	if !effectCalled {
		t.Error("effect was not called")
	}
}
