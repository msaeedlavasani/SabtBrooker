package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresCaseRepo implements CaseRepository
type PostgresCaseRepo struct {
	db *pgxpool.Pool
}

func NewPostgresCaseRepo(db *pgxpool.Pool) *PostgresCaseRepo {
	return &PostgresCaseRepo{db: db}
}

func (r *PostgresCaseRepo) ListByApplicant(ctx context.Context, applicantID uuid.UUID, offset, limit int) ([]Case, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, applicant_id, applicant_capacity::text, status::text,
		       province, city, district, village, address_detail,
		       legal_expert_id, survey_expert_id, proxy_verified,
		       map_tracking_code, claim_tracking_code, cert_tracking_code,
		       deadline_2years, applicant_deceased, deadline_5months,
		       created_at, updated_at, completed_at
		FROM cases WHERE applicant_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, applicantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list cases: %w", err)
	}
	defer rows.Close()
	return scanCases(rows)
}

func (r *PostgresCaseRepo) ListAll(ctx context.Context, offset, limit int) ([]Case, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, applicant_id, applicant_capacity::text, status::text,
		       province, city, district, village, address_detail,
		       legal_expert_id, survey_expert_id, proxy_verified,
		       map_tracking_code, claim_tracking_code, cert_tracking_code,
		       deadline_2years, applicant_deceased, deadline_5months,
		       created_at, updated_at, completed_at
		FROM cases ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list all cases: %w", err)
	}
	defer rows.Close()
	return scanCases(rows)
}

func (r *PostgresCaseRepo) Create(ctx context.Context, c *Case) (*Case, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO cases (applicant_id, applicant_capacity, province, city, district,
		                   village, address_detail, legal_expert_id, survey_expert_id)
		VALUES ($1, $2::applicant_capacity, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, applicant_id, applicant_capacity::text, status::text,
		          province, city, district, village, address_detail,
		          legal_expert_id, survey_expert_id, proxy_verified,
		          map_tracking_code, claim_tracking_code, cert_tracking_code,
		          deadline_2years, applicant_deceased, deadline_5months,
		          created_at, updated_at, completed_at
	`, c.ApplicantID, c.ApplicantCapacity, c.Province, c.City,
		c.District, c.Village, c.AddressDetail,
		c.LegalExpertID, c.SurveyExpertID,
	).Scan(
		&c.ID, &c.ApplicantID, &c.ApplicantCapacity, &c.Status,
		&c.Province, &c.City, &c.District, &c.Village, &c.AddressDetail,
		&c.LegalExpertID, &c.SurveyExpertID, &c.ProxyVerified,
		&c.MapTrackingCode, &c.ClaimTrackingCode, &c.CertTrackingCode,
		&c.Deadline2Years, &c.ApplicantDeceased, &c.Deadline5Months,
		&c.CreatedAt, &c.UpdatedAt, &c.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create case: %w", err)
	}
	return c, nil
}

func (r *PostgresCaseRepo) GetByID(ctx context.Context, id uuid.UUID) (*Case, error) {
	var c Case
	err := r.db.QueryRow(ctx, `
		SELECT id, applicant_id, applicant_capacity::text, status::text,
		       province, city, district, village, address_detail,
		       legal_expert_id, survey_expert_id, proxy_verified,
		       map_tracking_code, claim_tracking_code, cert_tracking_code,
		       deadline_2years, applicant_deceased, deadline_5months,
		       created_at, updated_at, completed_at
		FROM cases WHERE id = $1
	`, id).Scan(
		&c.ID, &c.ApplicantID, &c.ApplicantCapacity, &c.Status,
		&c.Province, &c.City, &c.District, &c.Village, &c.AddressDetail,
		&c.LegalExpertID, &c.SurveyExpertID, &c.ProxyVerified,
		&c.MapTrackingCode, &c.ClaimTrackingCode, &c.CertTrackingCode,
		&c.Deadline2Years, &c.ApplicantDeceased, &c.Deadline5Months,
		&c.CreatedAt, &c.UpdatedAt, &c.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}
	return &c, nil
}

func (r *PostgresCaseRepo) UpdateByID(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(fields))
	args := make([]interface{}, 0, len(fields)+1)
	argIdx := 1

	for field, value := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
		args = append(args, value)
		argIdx++
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE cases SET %s, updated_at = NOW() WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update case: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("case %s not found", id)
	}
	return nil
}

func (r *PostgresCaseRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE cases SET status = $1::case_status, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("case %s not found", id)
	}
	return nil
}

func (r *PostgresCaseRepo) UpdateCapacity(ctx context.Context, id uuid.UUID, capacity string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE cases SET applicant_capacity = $1::applicant_capacity, updated_at = NOW()
		WHERE id = $2
	`, capacity, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("case %s not found", id)
	}
	return nil
}

// helpers

func scanCases(rows pgx.Rows) ([]Case, error) {
	var cases []Case
	for rows.Next() {
		var c Case
		err := rows.Scan(
			&c.ID, &c.ApplicantID, &c.ApplicantCapacity, &c.Status,
			&c.Province, &c.City, &c.District, &c.Village, &c.AddressDetail,
			&c.LegalExpertID, &c.SurveyExpertID, &c.ProxyVerified,
			&c.MapTrackingCode, &c.ClaimTrackingCode, &c.CertTrackingCode,
			&c.Deadline2Years, &c.ApplicantDeceased, &c.Deadline5Months,
			&c.CreatedAt, &c.UpdatedAt, &c.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan case row: %w", err)
		}
		cases = append(cases, c)
	}
	return cases, nil
}
