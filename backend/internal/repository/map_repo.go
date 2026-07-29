package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresMapServiceRepo implements MapServiceRepository
type PostgresMapServiceRepo struct {
	db *pgxpool.Pool
}

func NewPostgresMapServiceRepo(db *pgxpool.Pool) *PostgresMapServiceRepo {
	return &PostgresMapServiceRepo{db: db}
}

func (r *PostgresMapServiceRepo) Create(ctx context.Context, caseID uuid.UUID) (*MapService, error) {
	var ms MapService
	err := r.db.QueryRow(ctx, `
		INSERT INTO map_services (case_id)
		VALUES ($1)
		RETURNING id, case_id, status::text, property_type, approx_area_sqm,
		          land_use, ownership_type, has_building, annex_count,
		          geo_latitude, geo_longitude, grant_access_to_others,
		          consent_granted_at, fieldwork_started_at, fieldwork_completed_at,
		          map_file_path, map_format, descriptive_table,
		          submitted_to_org_at, tracking_code, org_response_at,
		          org_rejection_reason, created_at, updated_at
	`, caseID).Scan(
		&ms.ID, &ms.CaseID, &ms.Status, &ms.PropertyType, &ms.ApproxAreaSqm,
		&ms.LandUse, &ms.OwnershipType, &ms.HasBuilding, &ms.AnnexCount,
		&ms.GeoLatitude, &ms.GeoLongitude, &ms.GrantAccessToOthers,
		&ms.ConsentGrantedAt, &ms.FieldworkStartedAt, &ms.FieldworkCompletedAt,
		&ms.MapFilePath, &ms.MapFormat, &ms.DescriptiveTable,
		&ms.SubmittedToOrgAt, &ms.TrackingCode, &ms.OrgResponseAt,
		&ms.OrgRejectionReason, &ms.CreatedAt, &ms.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create map service: %w", err)
	}
	return &ms, nil
}

func (r *PostgresMapServiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*MapService, error) {
	var ms MapService
	err := r.db.QueryRow(ctx, `
		SELECT id, case_id, status::text, property_type, approx_area_sqm,
		       land_use, ownership_type, has_building, annex_count,
		       geo_latitude, geo_longitude, grant_access_to_others,
		       consent_granted_at, fieldwork_started_at, fieldwork_completed_at,
		       map_file_path, map_format, descriptive_table,
		       submitted_to_org_at, tracking_code, org_response_at,
		       org_rejection_reason, created_at, updated_at
		FROM map_services WHERE id = $1
	`, id).Scan(
		&ms.ID, &ms.CaseID, &ms.Status, &ms.PropertyType, &ms.ApproxAreaSqm,
		&ms.LandUse, &ms.OwnershipType, &ms.HasBuilding, &ms.AnnexCount,
		&ms.GeoLatitude, &ms.GeoLongitude, &ms.GrantAccessToOthers,
		&ms.ConsentGrantedAt, &ms.FieldworkStartedAt, &ms.FieldworkCompletedAt,
		&ms.MapFilePath, &ms.MapFormat, &ms.DescriptiveTable,
		&ms.SubmittedToOrgAt, &ms.TrackingCode, &ms.OrgResponseAt,
		&ms.OrgRejectionReason, &ms.CreatedAt, &ms.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("map service not found: %w", err)
	}
	return &ms, nil
}

func (r *PostgresMapServiceRepo) RequestConsent(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE map_services SET updated_at = NOW()
		WHERE id = $1 AND status::text IN ('expert_assigned', 'pending_expert_assignment')
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("امکان درخواست رضایت در این وضعیت وجود ندارد")
	}
	return nil
}

func (r *PostgresMapServiceRepo) VerifyConsent(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE map_services SET consent_granted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND consent_granted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("رضایت قبلاً ثبت شده یا سرویس در وضعیت نامعتبر است")
	}
	return nil
}

func (r *PostgresMapServiceRepo) AssignExpert(ctx context.Context, id uuid.UUID, expertID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Update map_services status
	tag, err := tx.Exec(ctx, `
		UPDATE map_services SET status = 'expert_assigned', updated_at = NOW()
		WHERE id = $1 AND status::text = 'pending_expert_assignment'
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("سرویس نقشه در وضعیت منتظر تخصیص نیست")
	}

	// 2. Update case with expert ID
	_, err = tx.Exec(ctx, `
		UPDATE cases SET survey_expert_id = $2, updated_at = NOW()
		WHERE id = (SELECT case_id FROM map_services WHERE id = $1)
	`, id, expertID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresMapServiceRepo) SubmitFieldwork(ctx context.Context, id uuid.UUID, input SubmitFieldworkInput) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Update map_services
	tag, err := tx.Exec(ctx, `
		UPDATE map_services SET
			property_type = $2,
			approx_area_sqm = $3,
			land_use = $4,
			ownership_type = $5,
			has_building = $6,
			annex_count = $7,
			geo_latitude = $8,
			geo_longitude = $9,
			geo_location = ST_SetSRID(ST_MakePoint($9, $8), 4326),
			map_file_path = $10,
			map_format = $11,
			descriptive_table = $12,
			grant_access_to_others = $13,
			fieldwork_completed_at = NOW(),
			status = 'fieldwork_done',
			updated_at = NOW()
		WHERE id = $1 AND status::text = 'fieldwork_in_progress'
	`, id, input.PropertyType, input.ApproxAreaSqm, input.LandUse, input.OwnershipType,
		input.HasBuilding, input.AnnexCount, input.GeoLatitude, input.GeoLongitude,
		input.MapFilePath, input.MapFormat, input.DescriptiveTable, input.GrantAccessToOthers)

	if err != nil {
		return fmt.Errorf("failed to update map service: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("سرویس در وضعیت مناسب برای ثبت عملیات میدانی نیست")
	}

	// 2. Insert photos
	for _, p := range input.Photos {
		_, err = tx.Exec(ctx, `
			INSERT INTO map_photos (map_service_id, file_path, side, photo_latitude, photo_longitude, photo_location)
			VALUES ($1, $2, $3, $4, $5, ST_SetSRID(ST_MakePoint($5, $4), 4326))
		`, id, p.FilePath, p.Side, p.Latitude, p.Longitude)
		if err != nil {
			return fmt.Errorf("failed to insert map photo: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// SubmitToOrg simulates submitting to organization and auto-approves
// در نسخه نهایی: از Outbox Pattern برای ارسال واقعی به سازمان استفاده می‌شود
func (r *PostgresMapServiceRepo) SubmitToOrg(ctx context.Context, id uuid.UUID, status string) (uuid.UUID, string, error) {
	var caseID uuid.UUID
	trackingCode := "MAP-" + id.String()[:8]

	err := r.db.QueryRow(ctx, `
		UPDATE map_services SET
			status = 'approved',
			tracking_code = $2,
			org_response_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND status::text = $3
		RETURNING case_id
	`, id, trackingCode, status).Scan(&caseID)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to submit map to org: %w", err)
	}
	return caseID, trackingCode, nil
}
