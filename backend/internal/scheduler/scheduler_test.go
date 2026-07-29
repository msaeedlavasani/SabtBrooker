package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestSchedulerCreation verifies basic creation
func TestSchedulerCreation(t *testing.T) {
	s := New(nil, nil, 1*time.Hour)
	if s == nil {
		t.Fatal("scheduler should not be nil")
	}
	if s.interval != 1*time.Hour {
		t.Errorf("expected interval 1h, got %v", s.interval)
	}
	if len(s.jobs) != 3 {
		t.Errorf("expected 3 built-in handlers (deadline_2years, deadline_5months, otp_cleanup), got %d", len(s.jobs))
	}
}

// TestRegisterJob adds a custom handler
func TestRegisterJob(t *testing.T) {
	s := New(nil, nil, 1*time.Minute)

	called := false
	s.Register("custom_job", func(ctx context.Context, job ScheduledJob) error {
		called = true
		return nil
	})

	if len(s.jobs) != 4 {
		t.Errorf("expected 4 handlers after register, got %d", len(s.jobs))
	}

	handler, ok := s.jobs["custom_job"]
	if !ok {
		t.Fatal("custom job handler not found")
	}

	_ = handler(context.Background(), ScheduledJob{ID: uuid.New()})
	if !called {
		t.Error("custom handler was not called")
	}
}

// TestScheduleCreatesJob verifies Schedule inserts into DB
// This test requires a real PostgreSQL connection
func TestSchedule(t *testing.T) {
	// Integration test — needs database
	t.Skip("requires PostgreSQL connection")
}

// TestCancelDeadline marks remaining deadline jobs as skipped
// This test requires a real PostgreSQL connection
func TestCancelDeadline(t *testing.T) {
	// Integration test — needs database
	t.Skip("requires PostgreSQL connection")
}

// TestOTPCleanupHandler verifies the handler logic
func TestOTPCleanupHandler(t *testing.T) {
	// Unit test for handler logic — needs database
	t.Skip("requires PostgreSQL connection")
}

// TestDeadline2YearsHandler_AlreadyCompleted verifies race condition safety
func TestDeadline2YearsHandler_AlreadyCompleted(t *testing.T) {
	// Integration test — needs database to verify status check
	t.Skip("requires PostgreSQL connection")
}

// TestWithDB_ContextInjection verifies context helper
func TestWithDB_ContextInjection(t *testing.T) {
	ctx := context.Background()
	ctx = WithDB(ctx, nil)

	db := getDBFromCtx(ctx)
	// nil is expected since we injected nil
	if db != nil {
		t.Error("expected nil db from context")
	}
}

// TestWithDB_NilContext verifies empty context returns nil
func TestWithDB_NilContext(t *testing.T) {
	db := getDBFromCtx(context.Background())
	if db != nil {
		t.Error("expected nil from empty context")
	}
}
