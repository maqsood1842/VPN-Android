package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, s *Subscription) error {
	query := `
		INSERT INTO subscriptions (id, user_id, plan, status, platform, purchase_token, original_transaction_id, current_period_start, current_period_end, auto_renew, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		s.ID, s.UserID, s.Plan, s.Status, s.Platform, s.PurchaseToken,
		s.OriginalTransactionID, s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.AutoRenew, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	query := `
		SELECT id, user_id, plan, status, platform, purchase_token, original_transaction_id, current_period_start, current_period_end, auto_renew, created_at, updated_at
		FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1
	`
	s := &Subscription{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.Plan, &s.Status, &s.Platform, &s.PurchaseToken,
		&s.OriginalTransactionID, &s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	query := `
		SELECT id, user_id, plan, status, platform, purchase_token, original_transaction_id, current_period_start, current_period_end, auto_renew, created_at, updated_at
		FROM subscriptions WHERE user_id = $1 AND status = 'active' AND current_period_end > NOW()
		ORDER BY created_at DESC LIMIT 1
	`
	s := &Subscription{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.Plan, &s.Status, &s.Platform, &s.PurchaseToken,
		&s.OriginalTransactionID, &s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no active subscription")
		}
		return nil, err
	}
	return s, nil
}

func (r *Repository) Update(ctx context.Context, s *Subscription) error {
	query := `
		UPDATE subscriptions SET
			plan = $2, status = $3, platform = $4, purchase_token = $5,
			original_transaction_id = $6, current_period_start = $7,
			current_period_end = $8, auto_renew = $9, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		s.ID, s.Plan, s.Status, s.Platform, s.PurchaseToken,
		s.OriginalTransactionID, s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.AutoRenew,
	)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, userID uuid.UUID, status string, periodEnd *time.Time) error {
	query := `
		UPDATE subscriptions SET status = $2, current_period_end = $3, updated_at = NOW()
		WHERE user_id = $1 AND status = 'active'
	`
	_, err := r.db.Exec(ctx, query, userID, status, periodEnd)
	return err
}