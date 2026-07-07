package models

import "time"

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleCashier UserRole = "cashier"
	RoleKitchen UserRole = "kitchen"
)

type OrderSource string

const (
	SourceKiosk   OrderSource = "kiosk"
	SourceCashier OrderSource = "cashier"
	SourceMobile  OrderSource = "mobile"
	SourceOnline  OrderSource = "online"
)

type OrderStatus string

const (
	StatusDraft          OrderStatus = "draft"
	StatusPendingPayment OrderStatus = "pending_payment"
	StatusPaid           OrderStatus = "paid"
	StatusInPreparation  OrderStatus = "in_preparation"
	StatusReadyForPickup OrderStatus = "ready_for_pickup"
	StatusCompleted      OrderStatus = "completed"
	StatusCancelled      OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type PaymentProvider string

const (
	ProviderCash         PaymentProvider = "cash"
	ProviderCardTerminal PaymentProvider = "card_terminal"
	ProviderGCash        PaymentProvider = "gcash"
	ProviderMaya         PaymentProvider = "maya"
	ProviderPayMongo     PaymentProvider = "paymongo"
	ProviderStripe       PaymentProvider = "stripe"
)

type Store struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address" db:"address"`
	Timezone  string    `json:"timezone" db:"timezone"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

type User struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	ID        string    `json:"id" db:"id"`
	StoreID   string    `json:"store_id" db:"store_id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password_hash"`
	Role      UserRole  `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

type Category struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	ID        string    `json:"id" db:"id"`
	StoreID   string    `json:"store_id" db:"store_id"`
	Name      string    `json:"name" db:"name"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

type Product struct {
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ID          string    `json:"id" db:"id"`
	CategoryID  string    `json:"category_id" db:"category_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	BasePrice   int64     `json:"base_price" db:"base_price"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
}

type ModifierGroup struct {
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	MinSelection int       `json:"min_selection" db:"min_selection"`
	MaxSelection int       `json:"max_selection" db:"max_selection"`
}

type ModifierOption struct {
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	ID              string    `json:"id" db:"id"`
	ModifierGroupID string    `json:"modifier_group_id" db:"modifier_group_id"`
	Name            string    `json:"name" db:"name"`
	ExtraPrice      int64     `json:"extra_price" db:"extra_price"`
	IsDefault       bool      `json:"is_default" db:"is_default"`
}

// ModifierGroupWithOptions is a ModifierGroup along with its available options,
// as returned by queries that join modifier_groups to modifier_options.
type ModifierGroupWithOptions struct {
	Options []ModifierOption `json:"options"`
	ModifierGroup
}

type ProductModifierGroup struct {
	ProductID       string `json:"product_id" db:"product_id"`
	ModifierGroupID string `json:"modifier_group_id" db:"modifier_group_id"`
}

type Order struct {
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	ID            string        `json:"id" db:"id"`
	StoreID       string        `json:"store_id" db:"store_id"`
	Source        OrderSource   `json:"source" db:"source"`
	Status        OrderStatus   `json:"status" db:"status"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	OrderNumber   int           `json:"order_number" db:"order_number"`
	TotalAmount   int64         `json:"total_amount" db:"total_amount"`
}

type OrderItem struct {
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	ID                 string    `json:"id" db:"id"`
	OrderID            string    `json:"order_id" db:"order_id"`
	ProductID          string    `json:"product_id" db:"product_id"`
	ModifierOptionIDs  []string  `json:"-" db:"-"`
	Quantity           int       `json:"quantity" db:"quantity"`
	UnitPrice          int64     `json:"unit_price" db:"unit_price"`
	CalculatedSubtotal int64     `json:"calculated_subtotal" db:"calculated_subtotal"`
}

type OrderItemModifier struct {
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	ID               string    `json:"id" db:"id"`
	OrderItemID      string    `json:"order_item_id" db:"order_item_id"`
	ModifierOptionID string    `json:"modifier_option_id" db:"modifier_option_id"`
}

type Payment struct {
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	ID             string          `json:"id" db:"id"`
	OrderID        string          `json:"order_id" db:"order_id"`
	Provider       PaymentProvider `json:"provider" db:"provider"`
	Status         PaymentStatus   `json:"status" db:"status"`
	TransactionRef string          `json:"transaction_ref" db:"transaction_ref"`
	Amount         int64           `json:"amount" db:"amount"`
}
