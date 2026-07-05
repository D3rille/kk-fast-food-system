import { useEffect, useRef } from "react"
import type { OrderEvent } from "@/types/api"
import { getWsUrl } from "@/api/client"

interface UseOrdersWsOptions {
  onNewPending: () => void
  onConnectionChange?: (connected: boolean) => void
}

export function useOrdersWs({ onNewPending, onConnectionChange }: UseOrdersWsOptions) {
  const onNewPendingRef = useRef(onNewPending)
  const onConnectionChangeRef = useRef(onConnectionChange)
  onNewPendingRef.current = onNewPending
  onConnectionChangeRef.current = onConnectionChange

  useEffect(() => {
    let ws: WebSocket | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let destroyed = false

    function connect() {
      if (destroyed) return
      ws = new WebSocket(getWsUrl())

      ws.onopen = () => {
        onConnectionChangeRef.current?.(true)
      }

      ws.onmessage = (e: MessageEvent) => {
        try {
          const event = JSON.parse(e.data as string) as OrderEvent
          if (event.type === "order.status_changed" && event.status === "pending_payment") {
            onNewPendingRef.current()
          }
        } catch {
          // ignore malformed messages
        }
      }

      ws.onclose = () => {
        onConnectionChangeRef.current?.(false)
        if (!destroyed) {
          reconnectTimer = setTimeout(connect, 3000)
        }
      }

      ws.onerror = () => {
        ws?.close()
      }
    }

    connect()

    return () => {
      destroyed = true
      if (reconnectTimer) clearTimeout(reconnectTimer)
      ws?.close()
    }
  }, [])
}
