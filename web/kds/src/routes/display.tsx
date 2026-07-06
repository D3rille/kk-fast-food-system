import { useState, useEffect, useCallback } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { kitchenLogin } from "@/api/auth"
import { fetchKitchenOrders, fetchOrder } from "@/api/orders"
import { useKitchenWs } from "@/hooks/useKitchenWs"
import { DisplayCard } from "@/components/DisplayCard"
import type { OrderDetail, OrderEvent } from "@/types/api"
import { REMOVE_FROM_KITCHEN } from "@/types/api"

export const Route = createFileRoute("/display")({ component: ServingDisplay })

function ServingDisplay() {
  const queryClient = useQueryClient()
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loginError, setLoginError] = useState<string | null>(null)

  useEffect(() => {
    kitchenLogin()
      .then(() => setIsAuthenticated(true))
      .catch((e: Error) => setLoginError(e.message))
  }, [])

  const { data: orders = [] } = useQuery({
    queryKey: ["kitchen-orders"],
    queryFn: fetchKitchenOrders,
    enabled: isAuthenticated,
    refetchInterval: 30_000,
  })

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

  useKitchenWs({ onEvent: handleEvent })

  const sorted = [...orders].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  )

  const preparing = sorted.filter((o) => o.status === "in_preparation")
  const serving = sorted.filter((o) => o.status === "ready_for_pickup")

  return (
    <div className="flex flex-col min-h-svh bg-background">
      <header className="flex items-center justify-center px-6 py-4 border-b border-border bg-card shrink-0">
        <span className="text-2xl font-bold tracking-widest uppercase text-foreground">
          Order Status
        </span>
      </header>

      <main className="flex-1 p-6 overflow-auto">
        {loginError && (
          <div
            className="mb-4 rounded-md border px-4 py-3 text-sm"
            style={{
              borderColor: "rgba(239,68,68,0.4)",
              backgroundColor: "rgba(239,68,68,0.1)",
              color: "#fca5a5",
            }}
          >
            Unable to load orders.
          </div>
        )}

        <div className="grid grid-cols-2 gap-8 h-full">
          <section className="flex flex-col gap-4">
            <h2 className="text-center text-xl font-bold uppercase tracking-widest text-muted-foreground">
              Preparing
            </h2>
            <div className="grid grid-cols-2 gap-4">
              {preparing.map((order) => (
                <DisplayCard key={order.id} orderNumber={order.order_number} />
              ))}
            </div>
          </section>

          <section className="flex flex-col gap-4">
            <h2 className="text-center text-xl font-bold uppercase tracking-widest text-muted-foreground">
              Serving
            </h2>
            <div className="grid grid-cols-2 gap-4">
              {serving.map((order) => (
                <DisplayCard key={order.id} orderNumber={order.order_number} />
              ))}
            </div>
          </section>
        </div>
      </main>
    </div>
  )
}
