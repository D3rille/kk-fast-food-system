import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { IconArrowLeft, IconShoppingCart } from "@tabler/icons-react"
import { CartItem } from "@/components/cart/CartItem"
import { useCart } from "@/contexts/cart"
import { Button } from "@/components/ui/button"

export const Route = createFileRoute("/cart")({ component: CartScreen })

function formatPrice(centavos: number): string {
  return `₱${(centavos / 100).toFixed(2)}`
}

function CartScreen() {
  const navigate = useNavigate()
  const { items, total, dispatch } = useCart()

  if (items.length === 0) {
    return (
      <div className="flex min-h-svh flex-col items-center justify-center gap-6 px-8 text-center">
        <div className="flex h-24 w-24 items-center justify-center rounded-full bg-muted">
          <IconShoppingCart size={40} className="text-muted-foreground/50" />
        </div>
        <div>
          <h2 className="font-heading font-bold text-2xl">No items yet</h2>
          <p className="text-muted-foreground text-sm mt-1">
            Browse the menu and add something tasty.
          </p>
        </div>
        <Button size="lg" onClick={() => navigate({ to: "/menu" })}>
          Browse Menu
        </Button>
      </div>
    )
  }

  const itemCount = items.reduce((sum, i) => sum + i.quantity, 0)

  return (
    <div className="flex h-svh flex-col bg-background">
      <header className="flex items-center gap-4 px-6 py-4 border-b border-border shrink-0 bg-card">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate({ to: "/menu" })}
          className="h-11 w-11"
        >
          <IconArrowLeft size={20} />
        </Button>
        <div>
          <h1 className="font-heading font-bold text-xl">Your Order</h1>
          <p className="text-xs text-muted-foreground mt-0.5">
            {itemCount} {itemCount === 1 ? "item" : "items"}
          </p>
        </div>
      </header>

      <main className="flex-1 overflow-y-auto px-6">
        {items.map((item) => (
          <CartItem
            key={item.product.id}
            item={item}
            onIncrement={() =>
              dispatch({
                type: "UPDATE_QUANTITY",
                productId: item.product.id,
                quantity: item.quantity + 1,
              })
            }
            onDecrement={() =>
              dispatch({
                type: "UPDATE_QUANTITY",
                productId: item.product.id,
                quantity: item.quantity - 1,
              })
            }
            onRemove={() => dispatch({ type: "REMOVE_ITEM", productId: item.product.id })}
          />
        ))}
      </main>

      <footer className="shrink-0 border-t border-border bg-card px-6 py-5">
        <div className="flex items-baseline justify-between mb-5">
          <span className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
            Total
          </span>
          <span className="font-mono font-black text-3xl text-primary">
            {formatPrice(total)}
          </span>
        </div>
        <div className="flex gap-3">
          <Button
            variant="outline"
            size="lg"
            className="flex-none px-5 min-h-12"
            onClick={() => navigate({ to: "/menu" })}
          >
            + Add More
          </Button>
          <Button
            size="lg"
            className="flex-1 min-h-12 font-bold text-base"
            onClick={() => navigate({ to: "/payment" })}
          >
            Place Order
          </Button>
        </div>
      </footer>
    </div>
  )
}
