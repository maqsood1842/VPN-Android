package subscription

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID                     uuid.UUID  `json:"id"`
	UserID                 uuid.UUID  `json:"user_id"`
	Plan                   string     `json:"plan"`
	Status                 string     `json:"status"`
	Platform               string     `json:"platform"`
	PurchaseToken          string     `json:"purchase_token,omitempty"`
	OriginalTransactionID  string     `json:"original_transaction_id,omitempty"`
	CurrentPeriodStart     *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd       *time.Time `json:"current_period_end,omitempty"`
	AutoRenew              bool       `json:"auto_renew"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type SubscriptionRepository interface {
	Create(ctx context.Context, s *Subscription) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	Update(ctx context.Context, s *Subscription) error
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*Subscription, error)
}