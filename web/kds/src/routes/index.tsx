import { useState, useEffect, useCallback } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { kitchenLogin } from "@/api/auth"
import {
  fetchKitchenOrders,
  fetchOrder,
  startPreparation,
  markReady,
  completeOrder,
} from "@/api/orders"
import { useKitchenWs } from "@/hooks/useKitchenWs"
import { Header } from "@/components/Header"
import { OrderTicket } from "@/components/OrderTicket"
import type { Order, OrderDetail, OrderEvent } from "@/types/api"
import { REMOVE_FROM_KITCHEN } from "@/types/api"

export const Route = createFileRoute("/")({ component: StaffBoard })

function StaffBoard() {
  const queryClient = useQueryClient()
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loginError, setLoginError] = useState<string | null>(null)
  const [wsConnected, setWsConnected] = useState(false)
  const [now, setNow] = useState(() => Date.now())
  const [advancingId, setAdvancingId] = useState<string | null>(null)

  // Global 1s tick to drive elapsed timers on each ticket
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [])

  // Authenticate on mount with kitchen credentials
  useEffect(() => {
    kitchenLogin()
      .then(() => setIsAuthenticated(true))
      .catch((e: Error) => setLoginError(e.message))
  }, [])

  // Initial load: all active kitchen orders (paid, in_preparation, or ready_for_pickup)
  const { data: orders = [], isLoading } = useQuery({
    queryKey: ["kitchen-orders"],
    queryFn: fetchKitchenOrders,
    enabled: isAuthenticated,
    refetchInterval: 30_000,
  })

  // Handle WebSocket push events from the backend hub
  const handleEvent = useCallback(
    async (event: OrderEvent) => {
      if (event.type === "order.paid") {
        try {
          const order = await fetchOrder(event.order_id)
          queryClient.setQueryData<OrderDetail[]>(["kitchen-orders"], (old = []) => {
            if (old.find((o) => o.id === order.id)) return old
            return [...old, order]
          })
        } catch {
          // If fetch fails, the 30s poll will catch it
        }
      } else if (event.type === "order.status_changed") {
        queryClient.setQueryData<OrderDetail[]>(["kitchen-orders"], (old = []) => {
          if (REMOVE_FROM_KITCHEN.includes(event.status)) {
            return old.filter((o) => o.id !== event.order_id)
          }
          return old.map((o) =>
            o.id === event.order_id ? { ...o, status: event.status } : o,
          )
        })
      }
    },
    [queryClient],
  )

  useKitchenWs({ onEvent: handleEvent, onConnectionChange: setWsConnected })

  // advanceOrder only returns the order summary (no items) — merge in just the status
  // so the ticket doesn't lose its items/modifiers on the next render.
  function mergeStatus(updated: Order) {
    queryClient.setQueryData<OrderDetail[]>(["kitchen-orders"], (old = []) => {
      if (REMOVE_FROM_KITCHEN.includes(updated.status)) {
        return old.filter((o) => o.id !== updated.id)
      }
      return old.map((o) => (o.id === updated.id ? { ...o, status: updated.status } : o))
    })
  }

  const startPrepMutation = useMutation({
    mutationFn: startPreparation,
    onMutate: (id) => setAdvancingId(id),
    onSettled: () => setAdvancingId(null),
    onSuccess: mergeStatus,
  })

  const markReadyMutation = useMutation({
    mutationFn: markReady,
    onMutate: (id) => setAdvancingId(id),
    onSettled: () => setAdvancingId(null),
    onSuccess: mergeStatus,
  })

  const completeMutation = useMutation({
    mutationFn: completeOrder,
    onMutate: (id) => setAdvancingId(id),
    onSettled: () => setAdvancingId(null),
    onSuccess: mergeStatus,
  })

  // Sort FIFO — oldest first
  const sorted = [...orders].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  )

  const queued = sorted.filter((o) => o.status === "paid")
  const inPrep = sorted.filter((o) => o.status === "in_preparation")
  const ready = sorted.filter((o) => o.status === "ready_for_pickup")

  const columns = [
    { title: "Queued", orders: queued },
    { title: "In Prep", orders: inPrep },
    { title: "Ready for Pickup", orders: ready },
  ]

  return (
    <div className="flex flex-col min-h-svh bg-background">
      <Header
        connected={wsConnected}
        pending={queued.length}
        inPrep={inPrep.length}
        ready={ready.length}
      />

      <main className="flex-1 p-4 overflow-auto">
        {loginError && (
          <div
            className="mb-4 rounded-md border px-4 py-3 text-sm"
            style={{
              borderColor: "rgba(239,68,68,0.4)",
              backgroundColor: "rgba(239,68,68,0.1)",
              color: "#fca5a5",
            }}
          >
            Auth error: {loginError} — orders may not load.
          </div>
        )}

        {isLoading && (
          <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
            Loading orders…
          </div>
        )}

        {!isLoading && sorted.length === 0 && (
          <div className="flex flex-col items-center justify-center h-64 gap-3">
            <span className="text-5xl" role="img" aria-label="check">
              ✓
            </span>
            <p className="text-muted-foreground text-sm">All clear — no active orders.</p>
          </div>
        )}

        {sorted.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {columns.map((col) => (
              <div key={col.title} className="flex flex-col gap-3">
                <h2 className="text-xs font-bold uppercase tracking-widest text-muted-foreground px-1">
                  {col.title} ({col.orders.length})
                </h2>
                <div className="flex flex-col gap-4">
                  {col.orders.map((order) => (
                    <OrderTicket
                      key={order.id}
                      order={order}
                      now={now}
                      onStartPreparation={(id) => startPrepMutation.mutate(id)}
                      onMarkReady={(id) => markReadyMutation.mutate(id)}
                      onComplete={(id) => completeMutation.mutate(id)}
                      isAdvancing={advancingId === order.id}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
