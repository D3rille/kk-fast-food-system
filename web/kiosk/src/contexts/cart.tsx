import { createContext, useContext, useReducer } from "react"
import type { ReactNode } from "react"
import type { CartItem, Product, SelectedModifier } from "@/types/api"

type CartAction =
  | {
      type: "ADD_ITEM"
      product: Product
      selectedModifiers: SelectedModifier[]
      unitPrice: number
      quantity?: number
    }
  | { type: "REMOVE_ITEM"; itemId: string }
  | { type: "UPDATE_QUANTITY"; itemId: string; quantity: number }
  | { type: "CLEAR_CART" }

interface CartState {
  items: CartItem[]
}

// Two lines are the "same" cart entry only if they're the same product AND the same set of
// selected modifier options — a Regular fries and a Large fries must stay on separate lines.
function lineKey(productId: string, selectedModifiers: SelectedModifier[]): string {
  const optionIds = selectedModifiers
    .map((m) => m.optionId)
    .sort()
    .join(",")
  return `${productId}::${optionIds}`
}

function cartReducer(state: CartState, action: CartAction): CartState {
  switch (action.type) {
    case "ADD_ITEM": {
      const addQuantity = action.quantity ?? 1
      const key = lineKey(action.product.id, action.selectedModifiers)
      const existing = state.items.find(
        (i) => lineKey(i.product.id, i.selectedModifiers) === key,
      )
      if (existing) {
        return {
          items: state.items.map((i) =>
            i.id === existing.id ? { ...i, quantity: i.quantity + addQuantity } : i,
          ),
        }
      }
      const newItem: CartItem = {
        id: `${key}::${crypto.randomUUID()}`,
        product: action.product,
        quantity: addQuantity,
        selectedModifiers: action.selectedModifiers,
        unitPrice: action.unitPrice,
      }
      return { items: [...state.items, newItem] }
    }
    case "REMOVE_ITEM":
      return { items: state.items.filter((i) => i.id !== action.itemId) }
    case "UPDATE_QUANTITY": {
      if (action.quantity <= 0) {
        return { items: state.items.filter((i) => i.id !== action.itemId) }
      }
      return {
        items: state.items.map((i) =>
          i.id === action.itemId ? { ...i, quantity: action.quantity } : i,
        ),
      }
    }
    case "CLEAR_CART":
      return { items: [] }
    default:
      return state
  }
}

interface CartContextValue {
  items: CartItem[]
  total: number
  itemCount: number
  dispatch: React.Dispatch<CartAction>
}

const CartContext = createContext<CartContextValue | null>(null)

export function CartProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(cartReducer, { items: [] })

  const total = state.items.reduce((sum, item) => sum + item.unitPrice * item.quantity, 0)
  const itemCount = state.items.reduce((sum, item) => sum + item.quantity, 0)

  return (
    <CartContext.Provider value={{ items: state.items, total, itemCount, dispatch }}>
      {children}
    </CartContext.Provider>
  )
}

export function useCart(): CartContextValue {
  const ctx = useContext(CartContext)
  if (!ctx) throw new Error("useCart must be used within CartProvider")
  return ctx
}
