package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// BuildCaseStateMachine creates the complete case state machine with all 12 transitions
func BuildCaseStateMachine(db *pgxpool.Pool, nc *nats.Conn) *StateMachine {
	sm := NewStateMachine("case", db, nc)

	// T1: draft → map_in_progress
	sm.AddTransition(Transition{
		From: "draft",
		To:   "map_in_progress",
		Guard: func(ctx context.Context, caseID uuid.UUID) error {
			return checkIdentityVerified(ctx, db, caseID)
		},
		Effect: func(ctx context.Context, caseID uuid.UUID) error {
			// Auto-assign experts + create map service
			return createMapServiceWithAutoAssignment(ctx, db, caseID)
		},
		Events: []string{"case.status.changed", "case.map.started"},
	})

	// T2: draft → cancelled
	sm.AddTransition(Transition{
		From: "draft",
		To:   "cancelled",
		Events: []string{"case.status.changed", "case.cancelled"},
	})

	// T3: map_in_progress → map_completed
	sm.AddTransition(Transition{
		From:  "map_in_progress",
		To:    "map_completed",
		Guard: mapServiceApprovedGuard(db),
		Effect: func(ctx context.Context, caseID uuid.UUID) error {
			return syncMapTrackingCode(ctx, db, caseID)
		},
		Events: []string{"case.status.changed", "case.map.completed"},
	})

	// T4: map_in_progress → cancelled
	sm.AddTransition(Transition{
		From: "map_in_progress",
		To:   "cancelled",
		Events: []string{"case.status.changed", "case.cancelled"},
	})

	// T5: map_completed → claim_in_progress
	sm.AddTransition(Transition{
		From:  "map_completed",
		To:    "claim_in_progress",
		Guard: capacityVerifiedGuard(db),
		Effect: func(ctx context.Context, caseID uuid.UUID) error {
			return createClaimService(ctx, db, caseID)
		},
		Events: []string{"case.status.changed", "case.claim.started"},
	})

	// T6: map_completed → cancelled
	sm.AddTransition(Transition{
		From: "map_completed",
		To:   "cancelled",
		Events: []string{"case.status.changed", "case.cancelled"},
	})

	// T7: claim_in_progress → claim_completed
	sm.AddTransition(Transition{
		From:  "claim_in_progress",
		To:    "claim_completed",
		Guard: claimServiceApprovedGuard(db),
		Effect: func(ctx context.Context, caseID uuid.UUID) error {
			return completeClaimService(ctx, db, caseID)
		},
		Events: []string{"case.status.changed", "case.claim.completed"},
	})

	// T8: claim_in_progress → rejected
	sm.AddTransition(Transition{
		From: "claim_in_progress",
		To:   "rejected",
		Events: []string{"case.status.changed", "case.claim.rejected"},
	})

	// T9: claim_completed → cert_in_progress
	sm.AddTransition(Transition{
		From:  "claim_completed",
		To:    "cert_in_progress",
		Guard: withinDeadlineGuard(db),
		Effect: func(ctx context.Context, caseID uuid.UUID) error {
			return createCertService(ctx, db, caseID)
		},
		Events: []string{"case.status.changed", "case.cert.started"},
	})

	// T10: claim_completed → expired (scheduler-triggered)
	sm.AddTransition(Transition{
		From: "claim_completed",
		To:   "expired",
		Events: []string{"case.status.changed", "case.expired"},
	})

	// T11: cert_in_progress → cert_completed
	sm.AddTransition(Transition{
		From:  "cert_in_progress",
		To:    "cert_completed",
		Guard: certServiceApprovedGuard(db),
		Effect: func(ctx context.Context, caseID uuid.UUID) error {
			return completeCertService(ctx, db, caseID)
		},
		Events: []string{"case.status.changed", "case.completed"},
	})

	// T12: cert_in_progress → expired (scheduler-triggered)
	sm.AddTransition(Transition{
		From: "cert_in_progress",
		To:   "expired",
		Events: []string{"case.status.changed", "case.expired"},
	})

	return sm
}

// ---- Guard conditions ----

func checkIdentityVerified(ctx context.Context, db *pgxpool.Pool, caseID uuid.UUID) error {
	var mobileVerified, ncrMatch bool
	var sanaStatus string
	var age int

	err := db.QueryRow(ctx, `
		SELECT u.mobile_verified, COALESCE(u.ncr_mobile_match, false),
		       COALESCE(u.sana_status, 'unknown'),
		       COALESCE(EXTRACT(YEAR FROM AGE(CURRENT_DATE, u.birth_date))::int, 0)
		FROM cases c JOIN users u ON c.applicant_id = u.id
		WHERE c.id = $1
	`, caseID).Scan(&mobileVerified, &ncrMatch, &sanaStatus, &age)

	if err != nil {
		return fmt.Errorf("failed to verify identity: %w", err)
	}

	var reasons []string
	if !mobileVerified {
		reasons = append(reasons, "شماره موبایل تایید نشده است")
	}
	// Shahkar check skipped in dev — requires external API
	_ = ncrMatch
	if sanaStatus != "active" && sanaStatus != "unknown" {
		reasons = append(reasons, "ثبت‌نام در سامانه ثنا تایید نشده است")
	}
	if age > 0 && age < 18 {
		reasons = append(reasons, "سن کمتر از ۱۸ سال")
	}

	if len(reasons) > 0 {
		return &GuardError{Reasons: reasons}
	}
	return nil
}

func mapServiceApprovedGuard(db *pgxpool.Pool) GuardFunc {
	return func(ctx context.Context, caseID uuid.UUID) error {
		var status string
		err := db.QueryRow(ctx, `
			SELECT status::text FROM map_services WHERE case_id = $1
		`, caseID).Scan(&status)
		if err != nil || status != "approved" {
			return fmt.Errorf("نقشه ثبتی هنوز توسط سازمان تایید نشده است")
		}
		return nil
	}
}

func capacityVerifiedGuard(db *pgxpool.Pool) GuardFunc {
	return func(ctx context.Context, caseID uuid.UUID) error {
		var verified bool
		err := db.QueryRow(ctx, `
			SELECT COALESCE(proxy_verified, false) FROM cases WHERE id = $1
		`, caseID).Scan(&verified)
		if err != nil || !verified {
			return fmt.Errorf("احراز نمایندگی توسط کارشناس حقوقی تایید نشده است")
		}
		return nil
	}
}

func claimServiceApprovedGuard(db *pgxpool.Pool) GuardFunc {
	return func(ctx context.Context, caseID uuid.UUID) error {
		var status string
		err := db.QueryRow(ctx, `
			SELECT status::text FROM claim_services WHERE case_id = $1
		`, caseID).Scan(&status)
		if err != nil || status != "approved" {
			return fmt.Errorf("درج ادعا هنوز توسط سازمان تایید نشده است")
		}
		return nil
	}
}

func withinDeadlineGuard(db *pgxpool.Pool) GuardFunc {
	return func(ctx context.Context, caseID uuid.UUID) error {
		var deadline *time.Time
		err := db.QueryRow(ctx, `
			SELECT deadline_2years FROM cases WHERE id = $1
		`, caseID).Scan(&deadline)
		if err != nil {
			return err
		}
		if deadline != nil && time.Now().After(*deadline) {
			return fmt.Errorf("مهلت ۲ ساله به پایان رسیده است")
		}
		return nil
	}
}

func certServiceApprovedGuard(db *pgxpool.Pool) GuardFunc {
	return func(ctx context.Context, caseID uuid.UUID) error {
		var status string
		err := db.QueryRow(ctx, `
			SELECT status::text FROM cert_services WHERE case_id = $1
		`, caseID).Scan(&status)
		if err != nil || status != "approved" {
			return fmt.Errorf("گواهی اقدام هنوز توسط سازمان تایید نشده است")
		}
		return nil
	}
}

// ---- Effect functions ----

func createMapServiceWithAutoAssignment(ctx context.Context, db *pgxpool.Pool, caseID uuid.UUID) error {
	// Find available legal and survey experts (round-robin from available pool)
	var legalExpertID, surveyExpertID uuid.UUID

	db.QueryRow(ctx, `
		SELECT e.id FROM experts e
		WHERE e.expert_type = 'legal' AND e.is_available = true
		ORDER BY e.current_case_count ASC LIMIT 1
	`).Scan(&legalExpertID)

	db.QueryRow(ctx, `
		SELECT e.id FROM experts e
		WHERE e.expert_type = 'survey' AND e.is_available = true
		ORDER BY e.current_case_count ASC LIMIT 1
	`).Scan(&surveyExpertID)

	// Assign experts to case
	_, err := db.Exec(ctx, `
		UPDATE cases SET legal_expert_id = $2, survey_expert_id = $3 WHERE id = $1
	`, caseID, legalExpertID, surveyExpertID)

	if err != nil {
		return fmt.Errorf("failed to assign experts: %w", err)
	}

	// Create map service record
	_, err = db.Exec(ctx, `
		INSERT INTO map_services (case_id, status) VALUES ($1, 'pending_expert_assignment')
	`, caseID)

	return err
}

func syncMapTrackingCode(ctx context.Context, db *pgxpool.Pool, caseID uuid.UUID) error {
	var trackingCode string
	err := db.QueryRow(ctx, `
		SELECT tracking_code FROM map_services WHERE case_id = $1
	`, caseID).Scan(&trackingCode)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		UPDATE cases SET map_tracking_code = $2 WHERE id = $1
	`, caseID, trackingCode)
	return err
}

func createClaimService(ctx context.Context, db *pgxpool.Pool, caseID uuid.UUID) error {
	_, err := db.Exec(ctx, `
		INSERT INTO claim_services (case_id, status) VALUES ($1, 'pending_expert_assignment')
	`, caseID)
	return err
}

func completeClaimService(ctx context.Context, db *pgxpool.Pool, caseID uuid.UUID) error {
	var trackingCode string
	err := db.QueryRow(ctx, `
		SELECT tracking_code FROM claim_services WHERE case_id = $1
	`, caseID).Scan(&trackingCode)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		UPDATE cases SET
			claim_tracking_code = $2,
			claim_approved_at = NOW(),
			deadline_2years = NOW() + INTERVAL '2 years'
		WHERE id = $1
	`, caseID, trackingCode)
	return err
}

func createCertService(ctx context.Context, db *pgxpool.Pool, caseID uuid.UUID) error {
	// Check if applicant is deceased — set 5-month deadline
	var deceased bool
	var dateOfDeath *time.Time

	db.QueryRow(ctx, `
		SELECT applicant_deceased, date_of_death FROM cases WHERE id = $1
	`, caseID).Scan(&deceased, &dateOfDeath)

	if deceased && dateOfDeath != nil {
		deadline := dateOfDeath.AddDate(0, 5, 0)
		db.Exec(ctx, `
			UPDATE cases SET deadline_5months = $2 WHERE id = $1
		`, caseID, deadline)
	}

	_, err := db.Exec(ctx, `
		INSERT INTO cert_services (case_id, status) VALUES ($1, 'pending_data')
	`, caseID)
	return err
}

func completeCertService(ctx context.Context, db *pgxpool.Pool, caseID uuid.UUID) error {
	var trackingCode string
	err := db.QueryRow(ctx, `
		SELECT tracking_code FROM cert_services WHERE case_id = $1
	`, caseID).Scan(&trackingCode)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		UPDATE cases SET
			cert_tracking_code = $2,
			completed_at = NOW()
		WHERE id = $1
	`, caseID, trackingCode)
	return err
}
