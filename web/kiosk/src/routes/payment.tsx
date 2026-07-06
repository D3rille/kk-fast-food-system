import { useEffect, useState } from "react"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { IconArrowLeft, IconReceipt, IconAlertCircle, IconLoader2 } from "@tabler/icons-react"
import { createOrder, checkoutOrder } from "@/api/orders"
import { useCart } from "@/contexts/cart"
import { Button } from "@/components/ui/button"

export const Route = createFileRoute("/payment")({ component: PaymentScreen })

function formatPrice(centavos: number): string {
  return `₱${(centavos / 100).toFixed(2)}`
}

type PlaceState = "idle" | "processing" | "error"

function PaymentScreen() {
  const navigate = useNavigate()
  const { items, total } = useCart()
  const [state, setState] = useState<PlaceState>("idle")
  const [errorMessage, setErrorMessage] = useState<string>("")

  const storeId = import.meta.env.VITE_STORE_ID as string | undefined

  // Redirect only when the cart is empty on its own (e.g. a direct visit or a back/forward
  // navigation) — the cart is cleared from the confirmation screen, not here, so a successful
  // order never races this guard against the navigate to /confirmation below.
  useEffect(() => {
    if (items.length === 0) {
      navigate({ to: "/menu" })
    }
  }, [items.length, navigate])

  if (items.length === 0) {
    return null
  }

  async function handlePlaceOrder() {
    if (state === "processing") return
    setState("processing")
    setErrorMessage("")

    try {
      const orderItems = items.map((item) => ({
        product_id: item.product.id,
        quantity: item.quantity,
        unit_price: item.unitPrice,
        modifier_option_ids: item.selectedModifiers.map((m) => m.optionId),
      }))
      const order = await createOrder(storeId ?? "", total, orderItems)
      await checkoutOrder(order.id)
      navigate({
        to: "/confirmation",
        search: { orderNumber: order.order_number },
      })
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Something went wrong. Please try again."
      setErrorMessage(msg)
      setState("error")
    }
  }

  return (
    <div className="flex h-svh flex-col bg-background">
      <header className="flex items-center gap-4 px-6 py-4 border-b border-border shrink-0 bg-card">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate({ to: "/cart" })}
          disabled={state === "processing"}
          className="h-11 w-11"
        >
          <IconArrowLeft size={20} />
        </Button>
        <div>
          <h1 className="font-heading font-bold text-xl">Review Order</h1>
          <p className="text-xs text-muted-foreground mt-0.5">Confirm and place your order</p>
        </div>
      </header>

      <main className="flex flex-col flex-1 px-6 py-6 gap-6 overflow-y-auto">
        <div className="rounded-2xl bg-primary/8 border border-primary/20 p-6 text-center">
          <p className="text-sm font-semibold text-muted-foreground uppercase tracking-wide mb-1">
            Total Amount
          </p>
          <p className="font-mono font-black text-5xl text-primary leading-none">
            {formatPrice(total)}
          </p>
        </div>

        <div className="rounded-2xl border border-border bg-card p-5">
          <h2 className="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-3">
            Order Summary
          </h2>
          <div className="flex flex-col gap-3">
            {items.map((item) => (
              <div key={item.id} className="flex justify-between text-sm">
                <span className="truncate text-foreground/75">
                  {item.quantity}× {item.product.name}
                  {item.selectedModifiers.length > 0 && (
                    <span className="block text-xs text-muted-foreground/80 mt-0.5">
                      {item.selectedModifiers.map((m) => m.name).join(", ")}
                    </span>
                  )}
                </span>
                <span className="font-mono font-medium shrink-0 ml-4 tabular-nums">
                  {formatPrice(item.unitPrice * item.quantity)}
                </span>
              </div>
            ))}
          </div>
        </div>

        {state === "error" && (
          <div className="flex items-start gap-3 rounded-xl border border-destructive/30 bg-destructive/8 p-4 text-destructive">
            <IconAlertCircle size={18} className="shrink-0 mt-0.5" />
            <div>
              <p className="font-semibold text-sm">Something went wrong</p>
              <p className="text-xs mt-0.5 opacity-80">{errorMessage}</p>
            </div>
          </div>
        )}

        <div className="flex flex-col gap-2">
          <button
            onClick={handlePlaceOrder}
            disabled={state === "processing"}
            className="flex items-center gap-4 rounded-2xl border-2 border-primary bg-primary/5 p-5 text-left transition-all hover:bg-primary/10 active:scale-[0.99] disabled:opacity-60 disabled:cursor-not-allowed min-h-[76px]"
          >
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground">
              {state === "processing" ? (
                <IconLoader2 size={22} className="animate-spin" />
              ) : (
                <IconReceipt size={22} />
              )}
            </div>
            <div className="flex-1">
              <p className="font-bold text-base">Place Order</p>
              <p className="text-sm text-muted-foreground">
                {state === "processing"
                  ? "Placing your order…"
                  : "Pay at the cashier counter"}
              </p>
            </div>
          </button>
        </div>
      </main>
    </div>
  )
}
