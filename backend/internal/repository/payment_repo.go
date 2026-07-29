package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPaymentRepo struct {
	db *pgxpool.Pool
}

func NewPostgresPaymentRepo(db *pgxpool.Pool) *PostgresPaymentRepo {
	return &PostgresPaymentRepo{db: db}
}

func (r *PostgresPaymentRepo) GetActiveTariff(ctx context.Context, serviceType string) (*Tariff, error) {
	var t Tariff
	err := r.db.QueryRow(ctx, `
		SELECT id, service_type, max_amount, non_refundable, effective_from, version
		FROM tariffs WHERE service_type = $1 AND effective_to IS NULL
		LIMIT 1
	`, serviceType).Scan(&t.ID, &t.ServiceType, &t.MaxAmount, &t.NonRefundable, &t.EffectiveFrom, &t.Version)
	
	if err != nil {
		return nil, fmt.Errorf("active tariff not found for %s: %w", serviceType, err)
	}
	return &t, nil
}

func (r *PostgresPaymentRepo) CreatePayment(ctx context.Context, p *Payment) (*Payment, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO payments (case_id, service_type, amount, payment_type, status, psp_token, payment_url, tariff_version)
		VALUES ($1, $2, $3, $4::payment_type, $5::payment_status, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, p.CaseID, p.ServiceType, p.Amount, p.PaymentType, p.Status, p.PSPToken, p.PaymentURL, p.TariffVersion).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}
	return p, nil
}

func (r *PostgresPaymentRepo) GetPaymentByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	var p Payment
	err := r.db.QueryRow(ctx, `
		SELECT id, case_id, service_type, amount, payment_type::text, status::text, 
		       psp_token, psp_reference, payment_url, paid_at, tariff_version, created_at, updated_at
		FROM payments WHERE id = $1
	`, id).Scan(&p.ID, &p.CaseID, &p.ServiceType, &p.Amount, &p.PaymentType, &p.Status, 
		&p.PSPToken, &p.PSPReference, &p.PaymentURL, &p.PaidAt, &p.TariffVersion, &p.CreatedAt, &p.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}
	return &p, nil
}

func (r *PostgresPaymentRepo) GetPaymentByToken(ctx context.Context, token string) (*Payment, error) {
	var p Payment
	err := r.db.QueryRow(ctx, `
		SELECT id, case_id, service_type, amount, payment_type::text, status::text, 
		       psp_token, psp_reference, payment_url, paid_at, tariff_version, created_at, updated_at
		FROM payments WHERE psp_token = $1
	`, token).Scan(&p.ID, &p.CaseID, &p.ServiceType, &p.Amount, &p.PaymentType, &p.Status, 
		&p.PSPToken, &p.PSPReference, &p.PaymentURL, &p.PaidAt, &p.TariffVersion, &p.CreatedAt, &p.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("payment not found by token: %w", err)
	}
	return &p, nil
}

func (r *PostgresPaymentRepo) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string, ref string) error {
	query := `UPDATE payments SET status = $2::payment_status, psp_reference = $3, updated_at = NOW()`
	if status == "paid" {
		query += `, paid_at = NOW()`
	}
	query += ` WHERE id = $1`
	
	_, err := r.db.Exec(ctx, query, id, status, ref)
	return err
}
