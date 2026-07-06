import { useEffect, useState } from "react"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useAuth } from "@/contexts/auth"
import { fetchOrderDetail, payOrder, cancelOrder } from "@/api/orders"
import { Button } from "@/components/ui/button"
import type { OrderSource } from "@/types/api"

export const Route = createFileRoute("/orders/$orderId")({ component: OrderDetailPage })

function formatPrice(centavos: number): string {
  return `₱${(centavos / 100).toFixed(2)}`
}

const SOURCE_LABELS: Record<OrderSource, string> = {
  kiosk: "Kiosk",
  cashier: "Cashier",
  mobile: "Mobile",
  online: "Online",
}

const SOURCE_COLORS: Record<OrderSource, string> = {
  kiosk: "bg-purple-500/15 text-purple-400",
  cashier: "bg-blue-500/15 text-blue-400",
  mobile: "bg-teal-500/15 text-teal-400",
  online: "bg-orange-500/15 text-orange-400",
}

function OrderDetailPage() {
  const { orderId } = Route.useParams()
  const { token } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [cashInput, setCashInput] = useState("")
  const [payError, setPayError] = useState("")
  const [confirmingCancel, setConfirmingCancel] = useState(false)
  const [cancelError, setCancelError] = useState("")

  useEffect(() => {
    if (!token) {
      void navigate({ to: "/" })
    }
  }, [token, navigate])

  const {
    data: order,
    isLoading,
    error: fetchError,
  } = useQuery({
    queryKey: ["order", orderId],
    queryFn: () => fetchOrderDetail(orderId),
    enabled: !!token,
  })

  const payMutation = useMutation({
    mutationFn: () => payOrder(orderId, "cash"),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["orders", "pending"] })
      void queryClient.removeQueries({ queryKey: ["order", orderId] })
      void navigate({ to: "/orders" })
    },
    onError: (err: Error) => {
      setPayError(err.message)
    },
  })

  const cancelMutation = useMutation({
    mutationFn: () => cancelOrder(orderId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["orders", "pending"] })
      void queryClient.removeQueries({ queryKey: ["order", orderId] })
      void navigate({ to: "/orders" })
    },
    onError: (err: Error) => {
      setCancelError(err.message)
      setConfirmingCancel(false)
    },
  })

  if (!token) return null

  if (isLoading) {
    return (
      <div className="flex h-svh items-center justify-center text-sm text-muted-foreground">
        Loading order…
      </div>
    )
  }

  if (fetchError || !order) {
    return (
      <div className="flex h-svh flex-col items-center justify-center gap-4">
        <p className="text-sm text-destructive">Failed to load order.</p>
        <Button variant="outline" onClick={() => void navigate({ to: "/orders" })}>
          Back to queue
        </Button>
      </div>
    )
  }

  const totalPesos = order.total_amount / 100
  const cashPesos = parseFloat(cashInput) || 0
  const changePesos = cashPesos - totalPesos
  const canPay = cashPesos >= totalPesos && order.status === "pending_payment"
  const canCancel =
    order.status === "draft" || order.status === "pending_payment" || order.status === "paid"

  return (
    <div className="flex h-svh flex-col bg-background">
      <header className="flex items-center gap-4 border-b border-border bg-card px-6 py-4 shrink-0">
        <button
          onClick={() => void navigate({ to: "/orders" })}
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-border text-muted-foreground hover:text-foreground"
          aria-label="Back"
        >
          ←
        </button>
        <div className="flex flex-1 items-center gap-3">
          <span className="font-mono text-2xl font-black">#{order.order_number}</span>
          <span
            className={`rounded-full px-2.5 py-1 text-xs font-semibold ${SOURCE_COLORS[order.source]}`}
          >
            {SOURCE_LABELS[order.source]}
          </span>
        </div>
        <span className="text-sm text-muted-foreground">
          {new Date(order.created_at).toLocaleTimeString()}
        </span>
      </header>

      <div className="flex flex-1 flex-col gap-0 overflow-hidden md:flex-row">
        {/* Order items */}
        <div className="flex-1 overflow-y-auto p-6">
          <h2 className="mb-4 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Order Items
          </h2>

          {order.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">No item details available.</p>
          ) : (
            <div className="rounded-xl border border-border overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted/40">
                    <th className="px-4 py-3 text-left font-semibold text-muted-foreground">Item</th>
                    <th className="px-4 py-3 text-center font-semibold text-muted-foreground">Qty</th>
                    <th className="px-4 py-3 text-right font-semibold text-muted-foreground">Unit</th>
                    <th className="px-4 py-3 text-right font-semibold text-muted-foreground">Subtotal</th>
                  </tr>
                </thead>
                <tbody>
                  {order.items.map((item) => (
                    <tr key={item.id} className="border-b border-border last:border-0">
                      <td className="px-4 py-3 font-medium">
                        {item.product_name}
                        {item.modifiers.length > 0 && (
                          <span className="block text-xs font-normal text-muted-foreground mt-0.5">
                            {item.modifiers.map((m) => m.name).join(", ")}
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-center text-muted-foreground">{item.quantity}</td>
                      <td className="px-4 py-3 text-right font-mono tabular-nums text-muted-foreground">
                        {formatPrice(item.unit_price)}
                      </td>
                      <td className="px-4 py-3 text-right font-mono font-semibold tabular-nums">
                        {formatPrice(item.subtotal)}
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot>
                  <tr className="border-t-2 border-border bg-muted/40">
                    <td colSpan={3} className="px-4 py-3 font-bold">
                      Total
                    </td>
                    <td className="px-4 py-3 text-right font-mono font-black text-lg tabular-nums">
                      {formatPrice(order.total_amount)}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          )}
        </div>

        {/* Payment panel */}
        <div className="border-t border-border bg-card p-6 md:w-80 md:border-l md:border-t-0">
          <h2 className="mb-4 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Cash Payment
          </h2>

          <div className="flex flex-col gap-5">
            <div className="rounded-xl border border-border bg-background p-4 text-center">
              <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-1">
                Amount Due
              </p>
              <p className="font-mono font-black text-3xl text-primary">{formatPrice(order.total_amount)}</p>
            </div>

            <div className="flex flex-col gap-1.5">
              <label htmlFor="cash" className="text-sm font-medium">
                Cash Received (₱)
              </label>
              <input
                id="cash"
                type="number"
                min={totalPesos}
                step="0.01"
                placeholder={totalPesos.toFixed(2)}
                value={cashInput}
                onChange={(e) => {
                  setCashInput(e.target.value)
                  setPayError("")
                }}
                className="rounded-lg border border-input bg-background px-3 py-2 font-mono text-lg outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-ring"
              />
            </div>

            {cashInput && (
              <div className="rounded-xl border border-border p-4">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Cash received</span>
                  <span className="font-mono font-medium">₱{cashPesos.toFixed(2)}</span>
                </div>
                <div className="mt-2 flex justify-between">
                  <span className="font-semibold">Change</span>
                  <span
                    className={`font-mono font-bold text-lg ${changePesos >= 0 ? "text-green-500" : "text-destructive"}`}
                  >
                    {changePesos >= 0 ? `₱${changePesos.toFixed(2)}` : "Insufficient"}
                  </span>
                </div>
              </div>
            )}

            {payError && (
              <p className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {payError}
              </p>
            )}

            <Button
              onClick={() => {
                setPayError("")
                payMutation.mutate()
              }}
              disabled={!canPay || payMutation.isPending}
              className="w-full"
              size="lg"
            >
              {payMutation.isPending ? "Processing…" : "Confirm Payment"}
            </Button>

            {cancelError && (
              <p className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {cancelError}
              </p>
            )}

            {canCancel && !confirmingCancel && (
              <Button
                variant="outline"
                className="w-full"
                onClick={() => {
                  setCancelError("")
                  setConfirmingCancel(true)
                }}
              >
                Cancel this order
              </Button>
            )}

            {canCancel && confirmingCancel && (
              <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-4">
                <p className="text-sm font-medium text-center mb-3">Cancel this order?</p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    className="flex-1"
                    disabled={cancelMutation.isPending}
                    onClick={() => setConfirmingCancel(false)}
                  >
                    Keep Order
                  </Button>
                  <Button
                    variant="destructive"
                    className="flex-1"
                    disabled={cancelMutation.isPending}
                    onClick={() => cancelMutation.mutate()}
                  >
                    {cancelMutation.isPending ? "Cancelling…" : "Yes, Cancel"}
                  </Button>
                </div>
              </div>
            )}

            {order.status !== "pending_payment" && (
              <p className="text-center text-xs text-muted-foreground">
                This order is already {order.status.replace(/_/g, " ")}.
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
