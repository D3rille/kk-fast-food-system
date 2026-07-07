package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/repository"
	"github.com/D3rille/kk-fast-food-system/internal/service"
)

// --- Mocks ---

type mockOrderRepository struct {
	orders        map[string]*models.Order
	dailyCounters map[string]int
}

func newMockOrderRepository() *mockOrderRepository {
	return &mockOrderRepository{
		orders:        make(map[string]*models.Order),
		dailyCounters: make(map[string]int),
	}
}

func (m *mockOrderRepository) Create(_ context.Context, item *models.Order) error {
	m.orders[item.ID] = item
	return nil
}

func (m *mockOrderRepository) NextOrderNumber(_ context.Context, storeID string) (int, error) {
	key := storeID + "|" + time.Now().Format("2006-01-02")
	m.dailyCounters[key]++
	return m.dailyCounters[key], nil
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
	err      error
	result   *service.PaymentResult
	provider models.PaymentProvider
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

func TestOrderService_Create_OrderNumberIncrementsPerStore(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	for i, want := range []int{1, 2, 3} {
		order, err := svc.Create(context.Background(), &models.CreateOrderRequest{
			StoreID:     "store-1",
			Source:      models.SourceKiosk,
			TotalAmount: 10000,
		})
		if err != nil {
			t.Fatalf("create %d: expected no error, got %v", i, err)
		}
		if order.OrderNumber != want {
			t.Errorf("create %d: expected order_number %d, got %d", i, want, order.OrderNumber)
		}
	}

	// A different store's counter is independent and also starts at 1.
	otherStoreOrder, err := svc.Create(context.Background(), &models.CreateOrderRequest{
		StoreID:     "store-2",
		Source:      models.SourceKiosk,
		TotalAmount: 10000,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if otherStoreOrder.OrderNumber != 1 {
		t.Errorf("expected order_number 1 for a new store, got %d", otherStoreOrder.OrderNumber)
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

func TestOrderService_StartPreparation_Success(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	orderRepo.orders[order.ID].Status = models.StatusPaid

	updated, err := svc.StartPreparation(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != models.StatusInPreparation {
		t.Errorf("expected status in_preparation, got %q", updated.Status)
	}
	persisted := orderRepo.orders[order.ID]
	if persisted.Status != models.StatusInPreparation {
		t.Errorf("persisted status should be in_preparation, got %q", persisted.Status)
	}
}

func TestOrderService_StartPreparation_WrongState(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	// Order is still draft, not paid

	_, err := svc.StartPreparation(context.Background(), order.ID)
	if !errors.Is(err, service.ErrInvalidStateTransition) {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestOrderService_StartPreparation_OrderNotFound(t *testing.T) {
	svc := service.NewOrderService(newMockOrderRepository(), newMockPaymentRepository())

	_, err := svc.StartPreparation(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent order, got nil")
	}
}

func TestOrderService_MarkReady_Success(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	orderRepo.orders[order.ID].Status = models.StatusInPreparation

	updated, err := svc.MarkReady(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != models.StatusReadyForPickup {
		t.Errorf("expected status ready_for_pickup, got %q", updated.Status)
	}
	persisted := orderRepo.orders[order.ID]
	if persisted.Status != models.StatusReadyForPickup {
		t.Errorf("persisted status should be ready_for_pickup, got %q", persisted.Status)
	}
}

func TestOrderService_MarkReady_WrongState(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	orderRepo.orders[order.ID].Status = models.StatusPaid
	// Order is paid but not yet in_preparation

	_, err := svc.MarkReady(context.Background(), order.ID)
	if !errors.Is(err, service.ErrInvalidStateTransition) {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestOrderService_MarkReady_OrderNotFound(t *testing.T) {
	svc := service.NewOrderService(newMockOrderRepository(), newMockPaymentRepository())

	_, err := svc.MarkReady(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent order, got nil")
	}
}

func TestOrderService_Complete_Success(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	orderRepo.orders[order.ID].Status = models.StatusReadyForPickup

	updated, err := svc.Complete(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != models.StatusCompleted {
		t.Errorf("expected status completed, got %q", updated.Status)
	}
	persisted := orderRepo.orders[order.ID]
	if persisted.Status != models.StatusCompleted {
		t.Errorf("persisted status should be completed, got %q", persisted.Status)
	}
}

func TestOrderService_Complete_WrongState(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	orderRepo.orders[order.ID].Status = models.StatusInPreparation
	// Order is in_preparation but not yet ready_for_pickup

	_, err := svc.Complete(context.Background(), order.ID)
	if !errors.Is(err, service.ErrInvalidStateTransition) {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestOrderService_Complete_OrderNotFound(t *testing.T) {
	svc := service.NewOrderService(newMockOrderRepository(), newMockPaymentRepository())

	_, err := svc.Complete(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent order, got nil")
	}
}

func TestOrderService_Cancel_FromDraft(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)

	updated, err := svc.Cancel(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != models.StatusCancelled {
		t.Errorf("expected status cancelled, got %q", updated.Status)
	}
	persisted := orderRepo.orders[order.ID]
	if persisted.Status != models.StatusCancelled {
		t.Errorf("persisted status should be cancelled, got %q", persisted.Status)
	}
}

func TestOrderService_Cancel_FromPendingPayment(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	_, _ = svc.Checkout(context.Background(), order.ID)

	updated, err := svc.Cancel(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != models.StatusCancelled {
		t.Errorf("expected status cancelled, got %q", updated.Status)
	}
}

func TestOrderService_Cancel_FromPaid_RefundsPayment(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	order := seedDraftOrder(t, orderRepo, svc)
	orderRepo.orders[order.ID].Status = models.StatusPaid
	orderRepo.orders[order.ID].PaymentStatus = models.PaymentCompleted

	updated, err := svc.Cancel(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != models.StatusCancelled {
		t.Errorf("expected status cancelled, got %q", updated.Status)
	}
	if updated.PaymentStatus != models.PaymentRefunded {
		t.Errorf("expected payment_status refunded, got %q", updated.PaymentStatus)
	}
	persisted := orderRepo.orders[order.ID]
	if persisted.PaymentStatus != models.PaymentRefunded {
		t.Errorf("persisted payment_status should be refunded, got %q", persisted.PaymentStatus)
	}
}

func TestOrderService_Cancel_WrongState(t *testing.T) {
	orderRepo := newMockOrderRepository()
	paymentRepo := newMockPaymentRepository()
	svc := service.NewOrderService(orderRepo, paymentRepo)

	for _, status := range []models.OrderStatus{
		models.StatusInPreparation,
		models.StatusReadyForPickup,
		models.StatusCompleted,
		models.StatusCancelled,
	} {
		order := seedDraftOrder(t, orderRepo, svc)
		orderRepo.orders[order.ID].Status = status

		_, err := svc.Cancel(context.Background(), order.ID)
		if !errors.Is(err, service.ErrInvalidStateTransition) {
			t.Errorf("status %q: expected ErrInvalidStateTransition, got %v", status, err)
		}
	}
}

func TestOrderService_Cancel_OrderNotFound(t *testing.T) {
	svc := service.NewOrderService(newMockOrderRepository(), newMockPaymentRepository())

	_, err := svc.Cancel(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent order, got nil")
	}
}
