package service

import (
	"context"
	"fmt"

	"github.com/D3rille/kk-fast-food-system/internal/models"
)

// PaymentResult holds the outcome of a charge attempt.
type PaymentResult struct {
	Status         models.PaymentStatus
	TransactionRef string
}

// PaymentProvider abstracts a payment processing backend.
type PaymentProvider interface {
	// Provider returns the enum value used to record which provider was charged.
	Provider() models.PaymentProvider
	// Charge initiates payment for the given order and amount (in centavos).
	Charge(ctx context.Context, orderID string, amount int64) (*PaymentResult, error)
}

// cashProvider is a counter-cash stub that always succeeds immediately.
type cashProvider struct{}

// NewCashProvider returns a PaymentProvider that models cash collected at the counter.
func NewCashProvider() PaymentProvider {
	return &cashProvider{}
}

func (p *cashProvider) Provider() models.PaymentProvider {
	return models.ProviderCash
}

func (p *cashProvider) Charge(_ context.Context, orderID string, _ int64) (*PaymentResult, error) {
	return &PaymentResult{
		Status:         models.PaymentCompleted,
		TransactionRef: fmt.Sprintf("CASH-%s", orderID),
	}, nil
}
