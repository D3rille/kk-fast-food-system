import { useEffect, useState } from "react"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { useAuth } from "@/contexts/auth"
import { useOrdersWs } from "@/hooks/useOrdersWs"
import { fetchPendingOrders } from "@/api/orders"
import type { Order, OrderSource } from "@/types/api"

export const Route = createFileRoute("/orders/")({ component: OrdersPage })

function formatPrice(centavos: number): string {
  return `₱${(centavos / 100).toFixed(2)}`
}

function formatElapsed(secs: number): string {
  const m = Math.floor(secs / 60)
  const s = secs % 60
  return `${m}:${String(s).padStart(2, "0")}`
}

function ageColor(secs: number): string {
  if (secs < 300) return "#22c55e"
  if (secs < 600) return "#f59e0b"
  return "#ef4444"
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

function OrderCard({ order, now }: { order: Order; now: number }) {
  const navigate = useNavigate()
  const secs = Math.floor((now - new Date(order.created_at).getTime()) / 1000)
  const color = ageColor(secs)

  return (
    <button
      onClick={() => void navigate({ to: "/orders/$orderId", params: { orderId: order.id } })}
      className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-5 text-left transition-all hover:border-primary/40 hover:bg-card/80 active:scale-[0.99]"
      style={{ borderLeftColor: color, borderLeftWidth: 4 }}
    >
      <div className="flex items-start justify-between gap-3">
        <span className="font-mono font-black text-3xl leading-none">
          #{order.order_number}
        </span>
        <span
          className={`rounded-full px-2.5 py-1 text-xs font-semibold ${SOURCE_COLORS[order.source]}`}
        >
          {SOURCE_LABELS[order.source]}
        </span>
      </div>
      <div className="flex items-end justify-between">
        <span className="font-mono text-lg font-bold" style={{ color }}>
          {formatElapsed(secs)}
        </span>
        <span className="font-mono text-base font-semibold text-foreground">
          {formatPrice(order.total_amount)}
        </span>
      </div>
    </button>
  )
}

export function OrdersPage() {
  const { token, logout } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [now, setNow] = useState(Date.now())
  const [wsConnected, setWsConnected] = useState(false)

  useEffect(() => {
    if (!token) {
      void navigate({ to: "/" })
    }
  }, [token, navigate])

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [])

  useOrdersWs({
    onNewPending: () => {
      void queryClient.invalidateQueries({ queryKey: ["orders", "pending"] })
    },
    onConnectionChange: setWsConnected,
  })

  const { data: orders = [], isLoading } = useQuery({
    queryKey: ["orders", "pending"],
    queryFn: fetchPendingOrders,
    refetchInterval: 15_000,
    enabled: !!token,
  })

  if (!token) return null

  return (
    <div className="flex h-svh flex-col bg-background">
      <header className="flex items-center justify-between border-b border-border bg-card px-6 py-4 shrink-0">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-bold">Cashier POS</h1>
          {orders.length > 0 && (
            <span className="flex h-6 min-w-6 items-center justify-center rounded-full bg-amber-500/20 px-2 text-xs font-bold text-amber-400">
              {orders.length}
            </span>
          )}
        </div>
        <div className="flex items-center gap-4">
          <span className="flex items-center gap-1.5 text-xs">
            <span
              className={`h-2 w-2 rounded-full ${wsConnected ? "bg-green-500" : "bg-red-500"}`}
            />
            <span className="text-muted-foreground">{wsConnected ? "Live" : "Reconnecting…"}</span>
          </span>
          <button
            onClick={logout}
            className="text-xs text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
          >
            Sign out
          </button>
        </div>
      </header>

      <main className="flex-1 overflow-y-auto p-6">
        {isLoading ? (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            Loading orders…
          </div>
        ) : orders.length === 0 ? (
          <div className="flex h-full flex-col items-center justify-center gap-2 text-center">
            <p className="text-lg font-semibold">No pending orders</p>
            <p className="text-sm text-muted-foreground">Waiting for new kiosk orders…</p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {orders.map((order) => (
              <OrderCard key={order.id} order={order} now={now} />
            ))}
          </div>
        )}
      </main>

      <footer className="border-t border-border bg-card px-6 py-3 shrink-0">
        <p className="text-xs text-muted-foreground">
          {orders.length} order{orders.length !== 1 ? "s" : ""} pending payment ·{" "}
          {new Date().toLocaleTimeString()}
        </p>
      </footer>
    </div>
  )
}

