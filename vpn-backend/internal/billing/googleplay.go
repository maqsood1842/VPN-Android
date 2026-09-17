package billing

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
)

type GooglePlayValidator struct {
	service     *androidpublisher.Service
	packageName string
	productID   string
}

type SubscriptionPurchase struct {
	PackageName          string
	ProductID            string
	PurchaseToken        string
	StartTimeMillis      int64
	ExpiryTimeMillis     int64
	AutoRenewing         bool
	PriceCurrencyCode    string
	PriceAmountMicros    int64
	CountryCode          string
	PaymentState         int
	AcknowledgementState int
	OrderID              string
	PurchaseType         int
}

func NewGooglePlayValidator(ctx context.Context, credentialsJSON, packageName, productID string) (*GooglePlayValidator, error) {
	config, err := google.JWTConfigFromJSON([]byte(credentialsJSON), androidpublisher.AndroidpublisherScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	client := config.Client(ctx)
	service, err := androidpublisher.New(client)
	if err != nil {
		return nil, fmt.Errorf("create androidpublisher service: %w", err)
	}

	return &GooglePlayValidator{
		service:     service,
		packageName: packageName,
		productID:   productID,
	}, nil
}

func (v *GooglePlayValidator) ValidateSubscription(ctx context.Context, purchaseToken string) (*SubscriptionPurchase, error) {
	subscription, err := v.service.Purchases.Subscriptions.Get(v.packageName, v.productID, purchaseToken).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	paymentState := int64(0)
	if subscription.PaymentState != nil {
		paymentState = *subscription.PaymentState
	}

	purchaseType := int64(0)
	if subscription.PurchaseType != nil {
		purchaseType = *subscription.PurchaseType
	}

	purchase := &SubscriptionPurchase{
		PackageName:          v.packageName,
		ProductID:            v.productID,
		PurchaseToken:        purchaseToken,
		StartTimeMillis:      subscription.StartTimeMillis,
		ExpiryTimeMillis:     subscription.ExpiryTimeMillis,
		AutoRenewing:         subscription.AutoRenewing,
		PriceCurrencyCode:    subscription.PriceCurrencyCode,
		PriceAmountMicros:    subscription.PriceAmountMicros,
		CountryCode:          subscription.CountryCode,
		PaymentState:         int(paymentState),
		AcknowledgementState: int(subscription.AcknowledgementState),
		OrderID:              subscription.OrderId,
		PurchaseType:         int(purchaseType),
	}

	return purchase, nil
}

func (v *GooglePlayValidator) AcknowledgeSubscription(ctx context.Context, purchaseToken string) error {
	return v.service.Purchases.Subscriptions.Acknowledge(v.packageName, v.productID, purchaseToken, &androidpublisher.SubscriptionPurchasesAcknowledgeRequest{}).Context(ctx).Do()
}

func (v *GooglePlayValidator) IsSubscriptionActive(purchase *SubscriptionPurchase) bool {
	if purchase.PaymentState != 1 { // 1 = Payment received
		return false
	}

	expiryTime := time.UnixMilli(purchase.ExpiryTimeMillis)
	return time.Now().Before(expiryTime)
}

type PlayValidatorConfig struct {
	CredentialsJSON string `json:"credentials_json"`
	PackageName     string `json:"package_name"`
	ProductID       string `json:"product_id"`
}

func LoadConfigFromEnv() *PlayValidatorConfig {
	// In production, load from environment variables or secret manager
	return &PlayValidatorConfig{
		CredentialsJSON: "", // Set from env
		PackageName:     "com.vpn.app",
		ProductID:       "vpn_monthly_subscription",
	}
}