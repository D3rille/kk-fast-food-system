export type OrderStatus =
  | "draft"
  | "pending_payment"
  | "paid"
  | "in_preparation"
  | "ready_for_pickup"
  | "completed"
  | "cancelled"

export type OrderSource = "kiosk" | "cashier" | "mobile" | "online"

export interface Order {
  id: string
  store_id: string
  order_number: number
  source: OrderSource
  status: OrderStatus
  payment_status: string
  total_amount: number
  created_at: string
  updated_at: string
}

export interface OrderItemModifier {
  modifier_option_id: string
  group_name: string
  name: string
  extra_price: number
}

export interface OrderItem {
  id: string
  product_id: string
  product_name: string
  quantity: number
  unit_price: number
  subtotal: number
  created_at: string
  modifiers: OrderItemModifier[]
}

export interface OrderDetail extends Order {
  items: OrderItem[]
}

export interface OrderEvent {
  type: "order.created" | "order.paid" | "order.status_changed"
  order_id: string
  status: OrderStatus
  source: OrderSource
}

export const REMOVE_FROM_KITCHEN: OrderStatus[] = [
  "ready_for_pickup",
  "completed",
  "cancelled",
]
