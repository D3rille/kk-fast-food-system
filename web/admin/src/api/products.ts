import { apiFetch } from "./client"
import type { Product, ToggleAvailabilityResponse } from "@/types/api"

export async function fetchProducts(): Promise<Product[]> {
  return apiFetch<Product[]>("/api/v1/admin/menu/items")
}

export async function toggleAvailability(id: string): Promise<ToggleAvailabilityResponse> {
  return apiFetch<ToggleAvailabilityResponse>(`/api/v1/admin/menu/items/${id}/availability`, {
    method: "PATCH",
  })
}

export async function createProduct(formData: FormData): Promise<Product> {
  return apiFetch<Product>("/api/v1/admin/menu/items", {
    method: "POST",
    body: formData,
  })
}

export async function updateProduct(id: string, formData: FormData): Promise<Product> {
  return apiFetch<Product>(`/api/v1/admin/menu/items/${id}`, {
    method: "PUT",
    body: formData,
  })
}

export async function deleteProduct(id: string): Promise<void> {
  return apiFetch<void>(`/api/v1/admin/menu/items/${id}`, {
    method: "DELETE",
  })
}
