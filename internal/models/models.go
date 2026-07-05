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
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address" db:"address"`
	Timezone  string    `json:"timezone" db:"timezone"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID        string    `json:"id" db:"id"`
	StoreID   string    `json:"store_id" db:"store_id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password_hash"`
	Role      UserRole  `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Category struct {
	ID        string    `json:"id" db:"id"`
	StoreID   string    `json:"store_id" db:"store_id"`
	Name      string    `json:"name" db:"name"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Product struct {
	ID          string    `json:"id" db:"id"`
	CategoryID  string    `json:"category_id" db:"category_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	BasePrice   int64     `json:"base_price" db:"base_price"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ModifierGroup struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	MinSelection int       `json:"min_selection" db:"min_selection"`
	MaxSelection int       `json:"max_selection" db:"max_selection"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type ModifierOption struct {
	ID              string    `json:"id" db:"id"`
	ModifierGroupID string    `json:"modifier_group_id" db:"modifier_group_id"`
	Name            string    `json:"name" db:"name"`
	ExtraPrice      int64     `json:"extra_price" db:"extra_price"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type ProductModifierGroup struct {
	ProductID       string `json:"product_id" db:"product_id"`
	ModifierGroupID string `json:"modifier_group_id" db:"modifier_group_id"`
}

type Order struct {
	ID            string        `json:"id" db:"id"`
	StoreID       string        `json:"store_id" db:"store_id"`
	OrderNumber   int           `json:"order_number" db:"order_number"`
	Source        OrderSource   `json:"source" db:"source"`
	Status        OrderStatus   `json:"status" db:"status"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	TotalAmount   int64         `json:"total_amount" db:"total_amount"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID                 string    `json:"id" db:"id"`
	OrderID            string    `json:"order_id" db:"order_id"`
	ProductID          string    `json:"product_id" db:"product_id"`
	Quantity           int       `json:"quantity" db:"quantity"`
	UnitPrice          int64     `json:"unit_price" db:"unit_price"`
	CalculatedSubtotal int64     `json:"calculated_subtotal" db:"calculated_subtotal"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

type OrderItemModifier struct {
	ID               string    `json:"id" db:"id"`
	OrderItemID      string    `json:"order_item_id" db:"order_item_id"`
	ModifierOptionID string    `json:"modifier_option_id" db:"modifier_option_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type Payment struct {
	ID             string          `json:"id" db:"id"`
	OrderID        string          `json:"order_id" db:"order_id"`
	Provider       PaymentProvider `json:"provider" db:"provider"`
	Amount         int64           `json:"amount" db:"amount"`
	Status         PaymentStatus   `json:"status" db:"status"`
	TransactionRef string          `json:"transaction_ref" db:"transaction_ref"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}
