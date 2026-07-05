import type { Order, OrderStatus } from "@/types/api"
import { apiFetch } from "./client"

export async function fetchKitchenOrders(): Promise<Order[]> {
  const all = await apiFetch<Order[]>("/api/v1/orders")
  return all.filter((o) => o.status === "paid" || o.status === "in_preparation")
}

export async function fetchOrder(id: string): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${id}`)
}

export async function advanceOrder(id: string, status: OrderStatus): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${id}`, {
    method: "PUT",
    body: JSON.stringify({ id, status }),
  })
}
