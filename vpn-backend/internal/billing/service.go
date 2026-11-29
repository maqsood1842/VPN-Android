package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vpn/backend/internal/subscription"
)

type Service struct {
	validator  *GooglePlayValidator
	subRepo    *subscription.Repository
}

func NewService(validator *GooglePlayValidator, subRepo *subscription.Repository) *Service {
	return &Service{
		validator: validator,
		subRepo:   subRepo,
	}
}

type ValidatePurchaseRequest struct {
	PurchaseToken string `json:"purchase_token" binding:"required"`
	ProductID     string `json:"product_id" binding:"required"`
}

type ValidatePurchaseResponse struct {
	Success           bool   `json:"success"`
	SubscriptionID    string `json:"subscription_id,omitempty"`
	Status            string `json:"status"`
	ExpiryDate        string `json:"expiry_date,omitempty"`
	Error             string `json:"error,omitempty"`
}

func (s *Service) ValidateAndStorePurchase(ctx context.Context, userID uuid.UUID, req ValidatePurchaseRequest) (*ValidatePurchaseResponse, error) {
	if s.validator == nil {
		return &ValidatePurchaseResponse{
			Success: false,
			Error:   "Google Play validation not configured (missing credentials)",
		}, nil
	}

	purchase, err := s.validator.ValidateSubscription(ctx, req.PurchaseToken)
	if err != nil {
		return &ValidatePurchaseResponse{
			Success: false,
			Error:   fmt.Sprintf("validation failed: %v", err),
		}, nil
	}

	if !s.validator.IsSubscriptionActive(purchase) {
		return &ValidatePurchaseResponse{
			Success: false,
			Error:   "subscription not active",
		}, nil
	}

	// Acknowledge the purchase if not already acknowledged
	if purchase.AcknowledgementState != 1 { // 1 = Acknowledged
		if err := s.validator.AcknowledgeSubscription(ctx, req.PurchaseToken); err != nil {
			// Log but don't fail - acknowledgment can be retried
			fmt.Printf("Warning: failed to acknowledge purchase: %v\n", err)
		}
	}

	// Parse expiry time (milliseconds since epoch)
	expiryTime := time.UnixMilli(purchase.ExpiryTimeMillis)

	// Create or update subscription record
	sub := &subscription.Subscription{
		ID:                     uuid.New(),
		UserID:                 userID,
		Plan:                   purchase.ProductID,
		Status:                 "active",
		Platform:               "google_play",
		PurchaseToken:          req.PurchaseToken,
		OriginalTransactionID:  purchase.OrderID,
		CurrentPeriodStart:     nil, // Would parse from StartTimeMillis
		CurrentPeriodEnd:       &expiryTime,
		AutoRenew:              purchase.AutoRenewing,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return &ValidatePurchaseResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to store subscription: %v", err),
		}, nil
	}

	return &ValidatePurchaseResponse{
		Success:        true,
		SubscriptionID: sub.ID.String(),
		Status:         "active",
		ExpiryDate:     expiryTime.Format(time.RFC3339),
	}, nil
}

func (s *Service) RefreshSubscriptionStatus(ctx context.Context, userID uuid.UUID) error {
	if s.validator == nil {
		return fmt.Errorf("Google Play validation not configured (missing credentials)")
	}

	sub, err := s.subRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	purchase, err := s.validator.ValidateSubscription(ctx, sub.PurchaseToken)
	if err != nil {
		// If validation fails, mark as inactive
		s.subRepo.UpdateStatus(ctx, userID, "expired", nil)
		return err
	}

	if !s.validator.IsSubscriptionActive(purchase) {
		s.subRepo.UpdateStatus(ctx, userID, "expired", nil)
		return fmt.Errorf("subscription expired")
	}

	// Update expiry if extended
	expiryTime := time.UnixMilli(purchase.ExpiryTimeMillis)
	sub.CurrentPeriodEnd = &expiryTime
	sub.AutoRenew = purchase.AutoRenewing
	sub.UpdatedAt = time.Now()
	
	return s.subRepo.Update(ctx, sub)
}