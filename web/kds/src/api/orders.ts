import type { Order, OrderDetail } from "@/types/api"
import { apiFetch } from "./client"

// The list endpoint only returns order summaries (no items), so the kitchen display fetches
// each active order's detail (items + modifiers) individually to build its tickets.
export async function fetchKitchenOrders(): Promise<OrderDetail[]> {
  const all = await apiFetch<Order[]>("/api/v1/orders")
  const active = all.filter(
    (o) =>
      o.status === "paid" ||
      o.status === "in_preparation" ||
      o.status === "ready_for_pickup",
  )
  return Promise.all(active.map((o) => fetchOrder(o.id)))
}

export async function fetchOrder(id: string): Promise<OrderDetail> {
  return apiFetch<OrderDetail>(`/api/v1/orders/${id}`)
}

export async function startPreparation(id: string): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${id}/start-preparation`, { method: "POST" })
}

export async function markReady(id: string): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${id}/ready`, { method: "POST" })
}

export async function completeOrder(id: string): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${id}/complete`, { method: "POST" })
}
