package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/repository"
	"github.com/D3rille/kk-fast-food-system/internal/service"
)

// --- Mocks ---

type mockOrderRepository struct {
	orders map[string]*models.Order
}

func newMockOrderRepository() *mockOrderRepository {
	return &mockOrderRepository{orders: make(map[string]*models.Order)}
}

func (m *mockOrderRepository) Create(_ context.Context, item *models.Order) error {
	m.orders[item.ID] = item
	return nil
}

func (m *mockOrderRepository) GetByID(_ context.Context, id string) (*models.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, errors.New("order not found")
	}
	// return a copy so tests can observe mutations
	cp := *o
	return &cp, nil
}

func (m *mockOrderRepository) List(_ context.Context, _ string) ([]*models.Order, error) {
	items := make([]*models.Order, 0, len(m.orders))
	for _, o := range m.orders {
		cp := *o
		items = append(items, &cp)
	}
	return items, nil
}

func (m *mockOrderRepository) Update(_ context.Context, item *models.Order) error {
	if _, ok := m.orders[item.ID]; !ok {
		return errors.New("order not found")
	}
	cp := *item
	m.orders[item.ID] = &cp
	return nil
}

func (m *mockOrderRepository) UpdateStatus(_ context.Context, id string, status models.OrderStatus, paymentStatus models.PaymentStatus) error {
	o, ok := m.orders[id]
	if !ok {
		return errors.New("order not found")
	}
	o.Status = status
	o.PaymentStatus = paymentStatus
	return nil
}

func (m *mockOrderRepository) Delete(_ context.Context, id string) error {
	if _, ok := m.orders[id]; !ok {
		return errors.New("order not found")
	}
	delete(m.orders, id)
	return nil
}

// compile-time check
var _ repository.OrderRepository = (*mockOrderRepository)(nil)

type mockPaymentRepository struct {
	payments map[string]*models.Payment
}

func newMockPaymentRepository() *mockPaymentRepository {
	return &mockPaymentRepository{payments: make(map[string]*models.Payment)}
}

func (m *mockPaymentRepository) Create(_ context.Context, p *models.Payment) error {
	cp := *p
	m.payments[p.ID] = &cp
	return nil
}

func (m *mockPaymentRepository) UpdateStatus(_ context.Context, id string, status models.PaymentStatus, transactionRef string) error {
	p, ok := m.payments[id]
	if !ok {
		return errors.New("payment not found")
	}
	p.Status = status
	p.TransactionRef = transactionRef
	return nil
}

// compile-time check
var _ repository.PaymentRepository = (*mockPaymentRepository)(nil)

type mockPaymentProvider struct {
	provider models.PaymentProvider
	result   *service.PaymentResult
	err      error
}

func (m *mockPaymentProvider) Provider() models.PaymentProvider { return m.provider }
func (m *mockPaymentProvider) Charge(_ context.Context, _ string, _ int64) (*service.PaymentResult, error) {
	return m.result, m.err
}

// compile-time check
var _ service.PaymentProvider = (*mockPaymentProvider)(nil)

// --- Helpers ---

func seedDraftOrder(t *testing.T, repo *mockOrderRepository, svc service.OrderService) *models.Order {
	t.Helper()
	order, err := svc.Create(context.Background(), &models.CreateOrderRequest{
		StoreID:     "store-1",
		Source:      models.SourceKiosk,
		TotalAmount: 15000,
	})
	if err != nil {
		t.Fatalf("seed: Create failed: %v", err)
	}
	return order
}

// --- Tests ---

func TestOrderService_Create(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order, err := svc.Create(context.Background(), &models.CreateOrderRequest{
		StoreID:     "store-1",
		Source:      models.SourceKiosk,
		TotalAmount: 10000,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.ID == "" {
		t.Error("expected a generated ID")
	}
	if order.Status != models.StatusDraft {
		t.Errorf("expected status draft, got %q", order.Status)
	}
	if order.PaymentStatus != models.PaymentPending {
		t.Errorf("expected payment_status pending, got %q", order.PaymentStatus)
	}
	if order.TotalAmount != 10000 {
		t.Errorf("expected total_amount 10000, got %d", order.TotalAmount)
	}
}

func TestOrderService_Checkout_Success(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)

	updated, err := svc.Checkout(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != models.StatusPendingPayment {
		t.Errorf("expected status pending_payment, got %q", updated.Status)
	}
	// Verify persistence
	persisted := orderRepo.orders[order.ID]
	if persisted.Status != models.StatusPendingPayment {
		t.Errorf("persisted status should be pending_payment, got %q", persisted.Status)
	}
}

func TestOrderService_Checkout_WrongState(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	// Manually advance state to something other than draft
	orderRepo.orders[order.ID].Status = models.StatusPendingPayment

	_, err := svc.Checkout(context.Background(), order.ID)
	if !errors.Is(err, service.ErrInvalidStateTransition) {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestOrderService_Checkout_OrderNotFound(t *testing.T) {
	svc := service.NewOrderService(newMockOrderRepository(), newMockPaymentRepository())

	_, err := svc.Checkout(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent order, got nil")
	}
}

func TestOrderService_ProcessPayment_Success(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	_, _ = svc.Checkout(context.Background(), order.ID)

	provider := &mockPaymentProvider{
		provider: models.ProviderCash,
		result:   &service.PaymentResult{Status: models.PaymentCompleted, TransactionRef: "CASH-ref-123"},
	}

	payment, err := svc.ProcessPayment(context.Background(), order.ID, provider)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if payment.Status != models.PaymentCompleted {
		t.Errorf("expected payment status completed, got %q", payment.Status)
	}
	if payment.TransactionRef != "CASH-ref-123" {
		t.Errorf("expected transaction ref CASH-ref-123, got %q", payment.TransactionRef)
	}
	// Order should now be paid
	persistedOrder := orderRepo.orders[order.ID]
	if persistedOrder.Status != models.StatusPaid {
		t.Errorf("expected order status paid, got %q", persistedOrder.Status)
	}
	if persistedOrder.PaymentStatus != models.PaymentCompleted {
		t.Errorf("expected order payment_status completed, got %q", persistedOrder.PaymentStatus)
	}
}

func TestOrderService_ProcessPayment_WrongState(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	// Order is still draft (not checked out yet)
	order := seedDraftOrder(t, orderRepo, svc)

	provider := &mockPaymentProvider{
		provider: models.ProviderCash,
		result:   &service.PaymentResult{Status: models.PaymentCompleted},
	}

	_, err := svc.ProcessPayment(context.Background(), order.ID, provider)
	if !errors.Is(err, service.ErrInvalidStateTransition) {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestOrderService_ProcessPayment_ChargeFailure(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	_, _ = svc.Checkout(context.Background(), order.ID)

	chargeErr := errors.New("terminal offline")
	provider := &mockPaymentProvider{
		provider: models.ProviderCash,
		err:      chargeErr,
	}

	_, err := svc.ProcessPayment(context.Background(), order.ID, provider)
	if err == nil {
		t.Fatal("expected error from failed charge, got nil")
	}
	if !errors.Is(err, chargeErr) {
		t.Errorf("expected wrapped chargeErr, got %v", err)
	}

	// Order should remain in pending_payment with failed payment_status
	persistedOrder := orderRepo.orders[order.ID]
	if persistedOrder.Status != models.StatusPendingPayment {
		t.Errorf("expected order to remain pending_payment, got %q", persistedOrder.Status)
	}
	if persistedOrder.PaymentStatus != models.PaymentFailed {
		t.Errorf("expected order payment_status failed, got %q", persistedOrder.PaymentStatus)
	}

	// Payment record should be marked failed
	var failedPayment *models.Payment
	for _, p := range paymentRepo.payments {
		if p.OrderID == order.ID {
			failedPayment = p
			break
		}
	}
	if failedPayment == nil {
		t.Fatal("expected a payment record to be created")
	}
	if failedPayment.Status != models.PaymentFailed {
		t.Errorf("expected payment record status failed, got %q", failedPayment.Status)
	}
}
