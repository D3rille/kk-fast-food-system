import type { Order } from "@/types/api"
import { apiFetch } from "./client"

export interface OrderItemPayload {
  product_id: string
  quantity: number
  unit_price: number
}

export function createOrder(
  storeId: string,
  totalAmount: number,
  items: OrderItemPayload[],
): Promise<Order> {
  return apiFetch<Order>("/api/v1/orders", {
    method: "POST",
    body: JSON.stringify({
      store_id: storeId,
      source: "kiosk",
      total_amount: totalAmount,
      items,
    }),
  })
}

export function checkoutOrder(id: string): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${id}/checkout`, { method: "POST" })
}
