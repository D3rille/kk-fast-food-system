import type { Order, OrderDetail, OrderStatus } from "@/types/api"
import { apiFetch } from "./client"

// The list endpoint only returns order summaries (no items), so the kitchen display fetches
// each active order's detail (items + modifiers) individually to build its tickets.
export async function fetchKitchenOrders(): Promise<OrderDetail[]> {
  const all = await apiFetch<Order[]>("/api/v1/orders")
  const active = all.filter((o) => o.status === "paid" || o.status === "in_preparation")
  return Promise.all(active.map((o) => fetchOrder(o.id)))
}

export async function fetchOrder(id: string): Promise<OrderDetail> {
  return apiFetch<OrderDetail>(`/api/v1/orders/${id}`)
}

export async function advanceOrder(id: string, status: OrderStatus): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${id}`, {
    method: "PUT",
    body: JSON.stringify({ id, status }),
  })
}
