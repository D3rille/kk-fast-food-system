export type UserRole = "admin" | "manager" | "cashier" | "kitchen"

export interface User {
  id: string
  store_id: string
  username: string
  role: UserRole
  is_active: boolean
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user: User
}

export interface Product {
  id: string
  category_id: string
  name: string
  description: string
  base_price: number
  image_url: string
  is_available: boolean
}

export interface Category {
  id: string
  store_id: string
  name: string
  sort_order: number
  is_active: boolean
}

export interface ToggleAvailabilityResponse {
  id: string
  is_available: boolean
}
