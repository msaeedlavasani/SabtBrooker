package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresCertServiceRepo implements CertServiceRepository
type PostgresCertServiceRepo struct {
	db *pgxpool.Pool
}

func NewPostgresCertServiceRepo(db *pgxpool.Pool) *PostgresCertServiceRepo {
	return &PostgresCertServiceRepo{db: db}
}

func (r *PostgresCertServiceRepo) Create(ctx context.Context, caseID uuid.UUID) (*CertService, error) {
	var cs CertService
	err := r.db.QueryRow(ctx, `
		INSERT INTO cert_services (case_id)
		VALUES ($1)
		RETURNING id, case_id, status::text, claim_tracking_code,
		          claim_tracking_valid, consent_granted_at,
		          action_reference::text, action_type::text, action_date,
		          cert_image_path, cert_unique_id,
		          submitted_to_org_at, tracking_code, org_response_at,
		          org_rejection_reason, created_at, updated_at
	`, caseID).Scan(
		&cs.ID, &cs.CaseID, &cs.Status, &cs.ClaimTrackingCode,
		&cs.ClaimTrackingValid, &cs.ConsentGrantedAt,
		&cs.ActionReference, &cs.ActionType, &cs.ActionDate,
		&cs.CertImagePath, &cs.CertUniqueID,
		&cs.SubmittedToOrgAt, &cs.TrackingCode, &cs.OrgResponseAt,
		&cs.OrgRejectionReason, &cs.CreatedAt, &cs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert service: %w", err)
	}
	return &cs, nil
}

func (r *PostgresCertServiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*CertService, error) {
	var cs CertService
	err := r.db.QueryRow(ctx, `
		SELECT id, case_id, status::text, claim_tracking_code,
		       claim_tracking_valid, consent_granted_at,
		       action_reference::text, action_type::text, action_date,
		       cert_image_path, cert_unique_id,
		       submitted_to_org_at, tracking_code, org_response_at,
		       org_rejection_reason, created_at, updated_at
		FROM cert_services WHERE id = $1
	`, id).Scan(
		&cs.ID, &cs.CaseID, &cs.Status, &cs.ClaimTrackingCode,
		&cs.ClaimTrackingValid, &cs.ConsentGrantedAt,
		&cs.ActionReference, &cs.ActionType, &cs.ActionDate,
		&cs.CertImagePath, &cs.CertUniqueID,
		&cs.SubmittedToOrgAt, &cs.TrackingCode, &cs.OrgResponseAt,
		&cs.OrgRejectionReason, &cs.CreatedAt, &cs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("cert service not found: %w", err)
	}
	return &cs, nil
}

func (r *PostgresCertServiceRepo) RequestConsent(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE cert_services SET updated_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("امکان درخواست رضایت در این وضعیت وجود ندارد")
	}
	return nil
}

func (r *PostgresCertServiceRepo) VerifyConsent(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE cert_services SET consent_granted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND consent_granted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("رضایت قبلاً ثبت شده است")
	}
	return nil
}

func (r *PostgresCertServiceRepo) UpdateDetails(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	clauses := make([]string, 0)
	args := make([]interface{}, 0)
	idx := 1

	for field, value := range fields {
		switch field {
		case "action_reference", "action_type":
			clauses = append(clauses, fmt.Sprintf("%s = $%d::%s", field, idx, field))
		case "action_date":
			clauses = append(clauses, fmt.Sprintf("%s = $%d::date", field, idx))
		default:
			clauses = append(clauses, fmt.Sprintf("%s = $%d", field, idx))
		}
		args = append(args, value)
		idx++
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE cert_services SET %s, updated_at = NOW() WHERE id = $%d",
		joinClauses(clauses), idx)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update cert: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("cert %s not found", id)
	}
	return nil
}

// SubmitToOrg simulates submission and auto-approves
func (r *PostgresCertServiceRepo) SubmitToOrg(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error) {
	var caseID uuid.UUID
	trackingCode := "CERT-" + id.String()[:8]

	err := r.db.QueryRow(ctx, `
		UPDATE cert_services SET
			status = 'approved',
			tracking_code = $2,
			org_response_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		RETURNING case_id
	`, id, trackingCode).Scan(&caseID)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to submit cert to org: %w", err)
	}
	return caseID, trackingCode, nil
}
