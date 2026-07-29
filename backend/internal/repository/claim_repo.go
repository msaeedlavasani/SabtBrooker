package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresClaimServiceRepo implements ClaimServiceRepository
type PostgresClaimServiceRepo struct {
	db *pgxpool.Pool
}

func NewPostgresClaimServiceRepo(db *pgxpool.Pool) *PostgresClaimServiceRepo {
	return &PostgresClaimServiceRepo{db: db}
}

func (r *PostgresClaimServiceRepo) Create(ctx context.Context, caseID uuid.UUID) (*ClaimService, error) {
	var cs ClaimService
	err := r.db.QueryRow(ctx, `
		INSERT INTO claim_services (case_id)
		VALUES ($1)
		RETURNING id, case_id, status::text, map_tracking_code, map_tracking_valid,
		          false_claim_warning_sent, consent_granted_at,
		          claim_type::text, ownership_type::text,
		          main_plate_number, sub_plate_number, plate_section,
		          total_share, partial_share,
		          submitted_to_org_at, tracking_code, org_response_at,
		          org_rejection_reason, has_government_rights,
		          treasury_payment_ref, legal_advice_requested,
		          legal_advice_method, created_at, updated_at
	`, caseID).Scan(
		&cs.ID, &cs.CaseID, &cs.Status, &cs.MapTrackingCode, &cs.MapTrackingValid,
		&cs.FalseClaimWarningSent, &cs.ConsentGrantedAt,
		&cs.ClaimType, &cs.OwnershipType,
		&cs.MainPlateNumber, &cs.SubPlateNumber, &cs.PlateSection,
		&cs.TotalShare, &cs.PartialShare,
		&cs.SubmittedToOrgAt, &cs.TrackingCode, &cs.OrgResponseAt,
		&cs.OrgRejectionReason, &cs.HasGovernmentRights,
		&cs.TreasuryPaymentRef, &cs.LegalAdviceRequested,
		&cs.LegalAdviceMethod, &cs.CreatedAt, &cs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create claim service: %w", err)
	}
	return &cs, nil
}

func (r *PostgresClaimServiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*ClaimService, error) {
	var cs ClaimService
	err := r.db.QueryRow(ctx, `
		SELECT id, case_id, status::text, map_tracking_code, map_tracking_valid,
		       false_claim_warning_sent, consent_granted_at,
		       claim_type::text, ownership_type::text,
		       main_plate_number, sub_plate_number, plate_section,
		       total_share, partial_share,
		       submitted_to_org_at, tracking_code, org_response_at,
		       org_rejection_reason, has_government_rights,
		       treasury_payment_ref, legal_advice_requested,
		       legal_advice_method, created_at, updated_at
		FROM claim_services WHERE id = $1
	`, id).Scan(
		&cs.ID, &cs.CaseID, &cs.Status, &cs.MapTrackingCode, &cs.MapTrackingValid,
		&cs.FalseClaimWarningSent, &cs.ConsentGrantedAt,
		&cs.ClaimType, &cs.OwnershipType,
		&cs.MainPlateNumber, &cs.SubPlateNumber, &cs.PlateSection,
		&cs.TotalShare, &cs.PartialShare,
		&cs.SubmittedToOrgAt, &cs.TrackingCode, &cs.OrgResponseAt,
		&cs.OrgRejectionReason, &cs.HasGovernmentRights,
		&cs.TreasuryPaymentRef, &cs.LegalAdviceRequested,
		&cs.LegalAdviceMethod, &cs.CreatedAt, &cs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("claim service not found: %w", err)
	}
	return &cs, nil
}

func (r *PostgresClaimServiceRepo) RequestConsent(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE claim_services SET false_claim_warning_sent = true, updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("امکان درخواست رضایت در این وضعیت وجود ندارد")
	}
	return nil
}

func (r *PostgresClaimServiceRepo) VerifyConsent(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE claim_services SET consent_granted_at = NOW(), updated_at = NOW()
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

func (r *PostgresClaimServiceRepo) UpdateDetails(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(fields))
	args := make([]interface{}, 0, len(fields)+1)
	argIdx := 1

	for field, value := range fields {
		// handle enum types — نیاز به cast صریح
		switch v := value.(type) {
		case string:
			switch field {
			case "claim_type", "ownership_type":
				setClauses = append(setClauses, fmt.Sprintf("%s = $%d::%s", field, argIdx, field))
			default:
				setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
			}
			args = append(args, v)
		default:
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
			args = append(args, v)
		}
		argIdx++
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE claim_services SET %s, updated_at = NOW() WHERE id = $%d",
		joinClauses(setClauses), argIdx)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update claim: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("claim %s not found", id)
	}
	return nil
}

func (r *PostgresClaimServiceRepo) AddDocument(ctx context.Context, claimID uuid.UUID, docType, fileID, description string) (*ClaimDocument, error) {
	var d ClaimDocument
	err := r.db.QueryRow(ctx, `
		INSERT INTO claim_documents (claim_service_id, doc_type, file_path, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, claim_service_id, doc_type::text, file_path,
		          description, uploaded_at
	`, claimID, docType, fileID, description).Scan(
		&d.ID, &d.ClaimServiceID, &d.DocType, &d.FilePath, &d.Description, &d.UploadedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("خطا در افزودن مستند: %w", err)
	}
	return &d, nil
}

func (r *PostgresClaimServiceRepo) DeleteDocument(ctx context.Context, docID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM claim_documents WHERE id = $1 AND verified_at IS NULL
	`, docID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("مستند یافت نشد یا قبلاً تایید شده است")
	}
	return nil
}

func (r *PostgresClaimServiceRepo) ListDocuments(ctx context.Context, claimID uuid.UUID) ([]ClaimDocument, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, claim_service_id, doc_type::text, file_path,
		       description, COALESCE(verified_at IS NOT NULL, false),
		       COALESCE(verification_note, ''), uploaded_at
		FROM claim_documents WHERE claim_service_id = $1
		ORDER BY uploaded_at
	`, claimID)
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت مستندات: %w", err)
	}
	defer rows.Close()

	var docs []ClaimDocument
	for rows.Next() {
		var d ClaimDocument
		rows.Scan(&d.ID, &d.ClaimServiceID, &d.DocType, &d.FilePath,
			&d.Description, &d.Verified, &d.VerificationNote, &d.UploadedAt)
		docs = append(docs, d)
	}
	return docs, nil
}

func (r *PostgresClaimServiceRepo) VerifyDocument(ctx context.Context, docID uuid.UUID, claimID uuid.UUID, verified bool, note string) error {
	var verifiedAtClause string
	if verified {
		verifiedAtClause = "NOW()"
	} else {
		verifiedAtClause = "NULL"
	}

	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE claim_documents SET
			verified_by = (SELECT legal_expert_id FROM cases c
			               JOIN claim_services cs ON c.id = cs.case_id WHERE cs.id = $2),
			verified_at = %s,
			verification_note = $3
		WHERE id = $1
	`, verifiedAtClause), docID, claimID, note)
	if err != nil {
		return fmt.Errorf("خطا در ثبت تایید مستند: %w", err)
	}
	return nil
}

func (r *PostgresClaimServiceRepo) CountUnverifiedDocuments(ctx context.Context, claimID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM claim_documents
		WHERE claim_service_id = $1 AND verified_at IS NULL
	`, claimID).Scan(&count)
	return count, err
}

// SubmitToOrg simulates submission and auto-approves
func (r *PostgresClaimServiceRepo) SubmitToOrg(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error) {
	var caseID uuid.UUID
	trackingCode := "CLAIM-" + id.String()[:8]

	err := r.db.QueryRow(ctx, `
		UPDATE claim_services SET
			status = 'approved',
			tracking_code = $2,
			org_response_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		RETURNING case_id
	`, id, trackingCode).Scan(&caseID)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to submit claim to org: %w", err)
	}
	return caseID, trackingCode, nil
}

// helpers

func joinClauses(clauses []string) string {
	s := ""
	for i, c := range clauses {
		if i > 0 {
			s += ", "
		}
		s += c
	}
	return s
}
