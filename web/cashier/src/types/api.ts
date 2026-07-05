export type OrderStatus =
  | "draft"
  | "pending_payment"
  | "paid"
  | "in_preparation"
  | "ready_for_pickup"
  | "completed"
  | "cancelled"

export type OrderSource = "kiosk" | "cashier" | "mobile" | "online"

export type PaymentStatus = "pending" | "completed" | "failed" | "refunded"

export interface Order {
  id: string
  store_id: string
  order_number: number
  source: OrderSource
  status: OrderStatus
  payment_status: PaymentStatus
  total_amount: number // centavos
  created_at: string
  updated_at: string
}

export interface OrderItemModifier {
  modifier_option_id: string
  group_name: string
  name: string
  extra_price: number // centavos
}

export interface OrderItem {
  id: string
  product_id: string
  product_name: string
  quantity: number
  unit_price: number // centavos
  subtotal: number // centavos
  created_at: string
  modifiers: OrderItemModifier[]
}

export interface OrderDetail extends Order {
  items: OrderItem[]
}

export interface Payment {
  id: string
  order_id: string
  provider: string
  amount: number
  status: PaymentStatus
  transaction_ref: string
  created_at: string
  updated_at: string
}

export interface OrderEvent {
  type: "order.created" | "order.paid" | "order.status_changed"
  order_id: string
  status: OrderStatus
  source: OrderSource
}
