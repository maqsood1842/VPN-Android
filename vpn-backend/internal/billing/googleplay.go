package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
)

type GooglePlayValidator struct {
	service       *androidpublisher.Service
	packageName   string
	productID     string
}

type SubscriptionPurchase struct {
	PackageName           string `json:"packageName"`
	ProductID             string `json:"productId"`
	PurchaseToken         string `json:"purchaseToken"`
	StartTimeMillis       string `json:"startTimeMillis"`
	ExpiryTimeMillis      string `json:"expiryTimeMillis"`
	AutoRenewing          bool   `json:"autoRenewing"`
	PriceCurrencyCode     string `json:"priceCurrencyCode"`
	PriceAmountMicros     string `json:"priceAmountMicros"`
	CountryCode           string `json:"countryCode"`
	PaymentState          int    `json:"paymentState"`
	AcknowledgementState  int    `json:"acknowledgementState"`
	OrderID               string `json:"orderId"`
	PurchaseType          int    `json:"purchaseType"`
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

	purchase := &SubscriptionPurchase{
		PackageName:          subscription.PackageName,
		ProductID:            v.productID,
		PurchaseToken:        purchaseToken,
		StartTimeMillis:      subscription.StartTimeMillis,
		ExpiryTimeMillis:     subscription.ExpiryTimeMillis,
		AutoRenewing:         subscription.AutoRenewing,
		PriceCurrencyCode:    subscription.PriceCurrencyCode,
		PriceAmountMicros:    subscription.PriceAmountMicros,
		CountryCode:          subscription.CountryCode,
		PaymentState:         int(subscription.PaymentState),
		AcknowledgementState: int(subscription.AcknowledgementState),
		OrderID:              subscription.OrderId,
		PurchaseType:         int(subscription.PurchaseType),
	}

	return purchase, nil
}

func (v *GooglePlayValidator) AcknowledgeSubscription(ctx context.Context, purchaseToken string) error {
	return v.service.Purchases.Subscriptions.Acknowledge(v.packageName, v.productID, purchaseToken, &androidpublisher.SubscriptionPurchaseAcknowledgeRequest{}).Context(ctx).Do()
}

func (v *GooglePlayValidator) IsSubscriptionActive(purchase *SubscriptionPurchase) bool {
	if purchase.PaymentState != 1 { // 1 = Payment received
		return false
	}
	
	expiryTime, err := time.Parse("2006-01-02T15:04:05.000Z", purchase.ExpiryTimeMillis)
	if err != nil {
		// Try parsing as milliseconds since epoch
		return false
	}
	
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