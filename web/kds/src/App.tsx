import { useState, useEffect, useCallback } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { kitchenLogin } from "@/api/auth"
import { fetchKitchenOrders, fetchOrder, advanceOrder } from "@/api/orders"
import { useKitchenWs } from "@/hooks/useKitchenWs"
import { Header } from "@/components/Header"
import { OrderTicket } from "@/components/OrderTicket"
import type { Order, OrderEvent, OrderStatus } from "@/types/api"
import { REMOVE_FROM_KITCHEN } from "@/types/api"

export default function App() {
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

  // Initial load: all active kitchen orders (paid or in_preparation)
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
          queryClient.setQueryData<Order[]>(["kitchen-orders"], (old = []) => {
            if (old.find((o) => o.id === order.id)) return old
            return [...old, order]
          })
        } catch {
          // If fetch fails, the 30s poll will catch it
        }
      } else if (event.type === "order.status_changed") {
        queryClient.setQueryData<Order[]>(["kitchen-orders"], (old = []) => {
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

  // Advance an order's status and update the local cache immediately
  const advanceMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: OrderStatus }) =>
      advanceOrder(id, status),
    onMutate: ({ id }) => setAdvancingId(id),
    onSettled: () => setAdvancingId(null),
    onSuccess: (updated) => {
      queryClient.setQueryData<Order[]>(["kitchen-orders"], (old = []) => {
        if (REMOVE_FROM_KITCHEN.includes(updated.status)) {
          return old.filter((o) => o.id !== updated.id)
        }
        return old.map((o) => (o.id === updated.id ? updated : o))
      })
    },
  })

  function handleAdvance(id: string, status: OrderStatus) {
    advanceMutation.mutate({ id, status })
  }

  // Sort FIFO — oldest first
  const sorted = [...orders].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  )

  const pending = sorted.filter((o) => o.status === "paid").length
  const inPrep = sorted.filter((o) => o.status === "in_preparation").length

  return (
    <div className="flex flex-col min-h-svh bg-background">
      <Header connected={wsConnected} pending={pending} inPrep={inPrep} />

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
          <div
            className="grid gap-4"
            style={{ gridTemplateColumns: "repeat(auto-fill, minmax(260px, 1fr))" }}
          >
            {sorted.map((order) => (
              <OrderTicket
                key={order.id}
                order={order}
                now={now}
                onAdvance={handleAdvance}
                isAdvancing={advancingId === order.id}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
