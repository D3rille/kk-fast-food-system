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

export interface ModifierOption {
  id: string
  name: string
  extra_price: number
  is_default: boolean
}

export interface ModifierGroup {
  id: string
  name: string
  min_selection: number
  max_selection: number
  is_required: boolean
  options: ModifierOption[]
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

export interface SelectedModifier {
  groupId: string
  groupName: string
  optionId: string
  name: string
  extraPrice: number
}

// A cart line is keyed by `id`, not product.id — the same product with two different
// modifier configurations (e.g. Regular vs. Large fries) must appear as separate lines.
export interface CartItem {
  id: string
  product: Product
  quantity: number
  selectedModifiers: SelectedModifier[]
  unitPrice: number
}
