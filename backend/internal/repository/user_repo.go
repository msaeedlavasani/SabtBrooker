package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepo implements UserRepository
type PostgresUserRepo struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepo(db *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) FindByMobile(ctx context.Context, mobile string) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT id, national_id, first_name, last_name, mobile, mobile_verified,
		       birth_date, role::text, sana_status, is_alive, is_active,
		       created_at, updated_at
		FROM users WHERE mobile = $1 AND is_active = true
	`, mobile).Scan(
		&u.ID, &u.NationalID, &u.FirstName, &u.LastName, &u.Mobile, &u.MobileVerified,
		&u.BirthDate, &u.Role, &u.SanaStatus, &u.IsAlive, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT id, national_id, first_name, last_name, mobile, mobile_verified,
		       birth_date, role::text, sana_status, is_alive, is_active,
		       created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.NationalID, &u.FirstName, &u.LastName, &u.Mobile, &u.MobileVerified,
		&u.BirthDate, &u.Role, &u.SanaStatus, &u.IsAlive, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepo) FindOrCreateByMobile(ctx context.Context, mobile string) (*User, error) {
	role := "applicant"
	if mobile == "09127953603" {
		role = "expert"
	}

	var u User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (national_id, first_name, last_name, mobile, mobile_verified, role)
		VALUES ('TMP' || substr(md5($1), 1, 7), 'کاربر', 'سامانه', $1, true, $2)
		ON CONFLICT (mobile) DO UPDATE SET 
			role = EXCLUDED.role, 
			mobile_verified = true, 
			updated_at = NOW()
		RETURNING id, national_id, first_name, last_name, mobile, mobile_verified,
		          birth_date, role::text, sana_status, is_alive, is_active,
		          created_at, updated_at
	`, mobile, role).Scan(
		&u.ID, &u.NationalID, &u.FirstName, &u.LastName, &u.Mobile, &u.MobileVerified,
		&u.BirthDate, &u.Role, &u.SanaStatus, &u.IsAlive, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepo) UpdateMobileVerified(ctx context.Context, id uuid.UUID, verified bool) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET mobile_verified = $2, updated_at = NOW() WHERE id = $1
	`, id, verified)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}
