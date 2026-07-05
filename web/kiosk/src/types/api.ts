export interface Category {
  id: string
  store_id: string
  name: string
  sort_order: number
  is_active: boolean
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

export interface Order {
  id: string
  store_id: string
  order_number: number
  source: string
  status: string
  payment_status: string
  total_amount: number
  created_at: string
  updated_at: string
}

export interface Payment {
  id: string
  order_id: string
  provider: string
  amount: number
  status: string
  transaction_ref: string
  created_at: string
  updated_at: string
}

export interface CartItem {
  product: Product
  quantity: number
}
