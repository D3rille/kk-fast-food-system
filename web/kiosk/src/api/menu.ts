import type { Category, Product } from "@/types/api"
import { apiFetch } from "./client"

const storeId = import.meta.env.VITE_STORE_ID ?? "00000000-0000-0000-0000-000000000000"

export function getCategories(): Promise<Category[]> {
  return apiFetch<Category[]>("/api/v1/menu/categories", undefined, { store_id: storeId })
}

export function getProducts(categoryId?: string): Promise<Product[]> {
  return apiFetch<Product[]>("/api/v1/menu/items", undefined, { category_id: categoryId })
}
