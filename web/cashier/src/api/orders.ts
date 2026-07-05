import type { Order, OrderDetail, Payment } from "@/types/api"
import { apiFetch } from "./client"

export function fetchPendingOrders(): Promise<Order[]> {
  return apiFetch<Order[]>("/api/v1/orders?status=pending_payment")
}

export function fetchOrderDetail(id: string): Promise<OrderDetail> {
  return apiFetch<OrderDetail>(`/api/v1/orders/${id}`)
}

export function payOrder(id: string, provider: string): Promise<Payment> {
  return apiFetch<Payment>(`/api/v1/orders/${id}/pay`, {
    method: "POST",
    body: JSON.stringify({ provider }),
  })
}
