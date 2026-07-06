package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrOrderNotFound          = errors.New("order not found")
	ErrInvalidStateTransition = errors.New("invalid order state transition")
)

// OrderEvent is the payload broadcast to connected WebSocket clients on order state changes.
type OrderEvent struct {
	Type    string             `json:"type"`
	OrderID string             `json:"order_id"`
	Status  models.OrderStatus `json:"status"`
	Source  models.OrderSource `json:"source"`
}

// EventBroadcaster sends order events to real-time subscribers (e.g., the KDS WebSocket hub).
type EventBroadcaster interface {
	Broadcast(event OrderEvent)
}

// OrderService defines business operations for Order.
type OrderService interface {
	Create(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error)
	GetByID(ctx context.Context, id string) (*models.OrderDetailResponse, error)
	List(ctx context.Context, status string) ([]*models.Order, error)
	Update(ctx context.Context, id string, req *models.UpdateOrderRequest) (*models.Order, error)
	Delete(ctx context.Context, id string) error
	Checkout(ctx context.Context, orderID string) (*models.Order, error)
	ProcessPayment(ctx context.Context, orderID string, provider PaymentProvider) (*models.Payment, error)
}

type orderService struct {
	repo         repository.OrderRepository
	itemRepo     repository.OrderItemRepository
	paymentRepo  repository.PaymentRepository
	productRepo  repository.ProductRepository
	modifierRepo repository.ModifierRepository
	broadcaster  EventBroadcaster
}

// NewOrderService instantiates a new OrderService.
// An optional EventBroadcaster can be provided to fan out order events (e.g., to the KDS WebSocket hub).
func NewOrderService(repo repository.OrderRepository, paymentRepo repository.PaymentRepository, broadcaster ...EventBroadcaster) OrderService {
	var b EventBroadcaster
	if len(broadcaster) > 0 {
		b = broadcaster[0]
	}
	return &orderService{repo: repo, paymentRepo: paymentRepo, broadcaster: b}
}

// NewOrderServiceWithItems instantiates an OrderService that also persists order line items.
// productRepo and modifierRepo are used to recompute each item's price from the product's base
// price plus any selected modifier options, and to validate selections against each modifier
// group's min/max selection rules.
func NewOrderServiceWithItems(repo repository.OrderRepository, itemRepo repository.OrderItemRepository, paymentRepo repository.PaymentRepository, productRepo repository.ProductRepository, modifierRepo repository.ModifierRepository, broadcaster ...EventBroadcaster) OrderService {
	var b EventBroadcaster
	if len(broadcaster) > 0 {
		b = broadcaster[0]
	}
	return &orderService{repo: repo, itemRepo: itemRepo, paymentRepo: paymentRepo, productRepo: productRepo, modifierRepo: modifierRepo, broadcaster: b}
}

func (s *orderService) emit(event OrderEvent) {
	if s.broadcaster != nil {
		s.broadcaster.Broadcast(event)
	}
}

// Create always initialises the order as draft with pending payment status.
func (s *orderService) Create(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sequential ID: %w", err)
	}

	now := time.Now()
	item := &models.Order{
		ID:            id.String(),
		StoreID:       req.StoreID,
		Source:        req.Source,
		Status:        models.StatusDraft,
		PaymentStatus: models.PaymentPending,
		TotalAmount:   req.TotalAmount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("service failed to create order: %w", err)
	}

	if s.itemRepo != nil && len(req.Items) > 0 {
		orderItems := make([]*models.OrderItem, 0, len(req.Items))
		for _, ri := range req.Items {
			oid, err := uuid.NewV7()
			if err != nil {
				return nil, fmt.Errorf("failed to generate order item ID: %w", err)
			}

			unitPrice := ri.UnitPrice
			if s.productRepo != nil && s.modifierRepo != nil {
				unitPrice, err = s.priceItem(ctx, ri.ProductID, ri.ModifierOptionIDs)
				if err != nil {
					return nil, err
				}
			}

			orderItems = append(orderItems, &models.OrderItem{
				ID:                 oid.String(),
				OrderID:            item.ID,
				ProductID:          ri.ProductID,
				Quantity:           ri.Quantity,
				UnitPrice:          unitPrice,
				CalculatedSubtotal: unitPrice * int64(ri.Quantity),
				CreatedAt:          now,
				ModifierOptionIDs:  ri.ModifierOptionIDs,
			})
		}
		if err := s.itemRepo.CreateBatch(ctx, orderItems); err != nil {
			return nil, fmt.Errorf("service failed to create order items: %w", err)
		}
	}

	s.emit(OrderEvent{Type: "order.created", OrderID: item.ID, Status: item.Status, Source: item.Source})
	return item, nil
}

// priceItem recomputes an order item's unit price from the product's current base price plus
// the extra price of its selected modifier options, after validating the selections satisfy
// every attached modifier group's min/max selection rules. This keeps pricing authoritative on
// the server rather than trusting the client-supplied unit price.
func (s *orderService) priceItem(ctx context.Context, productID string, selectedOptionIDs []string) (int64, error) {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return 0, fmt.Errorf("service failed to price order item: %w", err)
	}
	groups, err := s.modifierRepo.ListGroupsForProduct(ctx, productID)
	if err != nil {
		return 0, fmt.Errorf("service failed to load modifier groups: %w", err)
	}
	extra, err := ValidateAndPriceSelections(groups, selectedOptionIDs)
	if err != nil {
		return 0, err
	}
	return product.BasePrice + extra, nil
}

func (s *orderService) GetByID(ctx context.Context, id string) (*models.OrderDetailResponse, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service failed to find order: %w", err)
	}

	detail := &models.OrderDetailResponse{
		OrderResponse: models.OrderResponse{
			ID:            order.ID,
			StoreID:       order.StoreID,
			OrderNumber:   order.OrderNumber,
			Source:        order.Source,
			Status:        order.Status,
			PaymentStatus: order.PaymentStatus,
			TotalAmount:   order.TotalAmount,
			CreatedAt:     order.CreatedAt,
			UpdatedAt:     order.UpdatedAt,
		},
		Items: []models.OrderItemResponse{},
	}

	if s.itemRepo != nil {
		items, err := s.itemRepo.GetByOrderIDWithProducts(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("service failed to fetch order items: %w", err)
		}
		detail.Items = items
	}

	return detail, nil
}

func (s *orderService) List(ctx context.Context, status string) ([]*models.Order, error) {
	items, err := s.repo.List(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("service failed to list order: %w", err)
	}
	return items, nil
}

func (s *orderService) Update(ctx context.Context, id string, req *models.UpdateOrderRequest) (*models.Order, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service failed to find order for update: %w", err)
	}

	item.StoreID = req.StoreID
	item.OrderNumber = req.OrderNumber
	item.Source = req.Source
	item.Status = req.Status
	item.PaymentStatus = req.PaymentStatus
	item.TotalAmount = req.TotalAmount
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("service failed to update order: %w", err)
	}
	return item, nil
}

func (s *orderService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service failed to delete order: %w", err)
	}
	return nil
}

// Checkout transitions an order from draft → pending_payment.
func (s *orderService) Checkout(ctx context.Context, orderID string) (*models.Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("checkout: %w: %w", ErrOrderNotFound, err)
	}
	if order.Status != models.StatusDraft {
		return nil, fmt.Errorf("checkout: order %s is in state %q: %w", orderID, order.Status, ErrInvalidStateTransition)
	}
	if err := s.repo.UpdateStatus(ctx, orderID, models.StatusPendingPayment, models.PaymentPending); err != nil {
		return nil, fmt.Errorf("checkout: failed to update order status: %w", err)
	}
	order.Status = models.StatusPendingPayment
	s.emit(OrderEvent{Type: "order.status_changed", OrderID: order.ID, Status: order.Status, Source: order.Source})
	return order, nil
}

// ProcessPayment charges the given provider and transitions the order from pending_payment → paid.
// If the charge fails the payment record is marked failed and the order remains in pending_payment.
func (s *orderService) ProcessPayment(ctx context.Context, orderID string, provider PaymentProvider) (*models.Payment, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("process payment: %w: %w", ErrOrderNotFound, err)
	}
	if order.Status != models.StatusPendingPayment {
		return nil, fmt.Errorf("process payment: order %s is in state %q: %w", orderID, order.Status, ErrInvalidStateTransition)
	}

	paymentID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("process payment: failed to generate payment ID: %w", err)
	}
	now := time.Now()
	payment := &models.Payment{
		ID:        paymentID.String(),
		OrderID:   orderID,
		Provider:  provider.Provider(),
		Amount:    order.TotalAmount,
		Status:    models.PaymentPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err = s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("process payment: failed to record payment: %w", err)
	}

	result, err := provider.Charge(ctx, orderID, order.TotalAmount)
	if err != nil {
		_ = s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentFailed, "")
		_ = s.repo.UpdateStatus(ctx, orderID, models.StatusPendingPayment, models.PaymentFailed)
		return nil, fmt.Errorf("process payment: charge failed: %w", err)
	}

	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, result.Status, result.TransactionRef); err != nil {
		return nil, fmt.Errorf("process payment: failed to update payment record: %w", err)
	}
	if err := s.repo.UpdateStatus(ctx, orderID, models.StatusPaid, models.PaymentCompleted); err != nil {
		return nil, fmt.Errorf("process payment: failed to transition order to paid: %w", err)
	}
	s.emit(OrderEvent{Type: "order.paid", OrderID: order.ID, Status: models.StatusPaid, Source: order.Source})

	payment.Status = result.Status
	payment.TransactionRef = result.TransactionRef
	return payment, nil
}
