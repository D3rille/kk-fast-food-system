import { apiFetch } from "./client"
import type { Category } from "@/types/api"

const storeId = import.meta.env.VITE_STORE_ID ?? "00000000-0000-0000-0000-000000000000"

export async function fetchCategories(): Promise<Category[]> {
  return apiFetch<Category[]>("/api/v1/admin/menu/categories", undefined, {store_id:storeId })
}
