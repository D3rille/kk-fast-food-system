import { useEffect, useLayoutEffect, useRef } from "react"
import type { OrderEvent } from "@/types/api"
import { getWsUrl } from "@/api/client"

interface UseKitchenWsOptions {
  onEvent: (event: OrderEvent) => void
  onConnectionChange?: (connected: boolean) => void
}

export function useKitchenWs({ onEvent, onConnectionChange }: UseKitchenWsOptions) {
  const onEventRef = useRef(onEvent)
  const onConnectionChangeRef = useRef(onConnectionChange)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const unmountedRef = useRef(false)

  useLayoutEffect(() => {
    onEventRef.current = onEvent
    onConnectionChangeRef.current = onConnectionChange
  })

  useEffect(() => {
    unmountedRef.current = false

    function connect() {
      if (unmountedRef.current) return

      const ws = new WebSocket(getWsUrl())
      wsRef.current = ws

      ws.onopen = () => {
        onConnectionChangeRef.current?.(true)
      }

      ws.onmessage = (event: MessageEvent<string>) => {
        try {
          const data = JSON.parse(event.data) as OrderEvent
          onEventRef.current(data)
        } catch {
          // ignore malformed messages
        }
      }

      ws.onclose = () => {
        onConnectionChangeRef.current?.(false)
        if (!unmountedRef.current) {
          reconnectTimerRef.current = setTimeout(connect, 3000)
        }
      }

      ws.onerror = () => {
        ws.close()
      }
    }

    connect()

    return () => {
      unmountedRef.current = true
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current)
      wsRef.current?.close()
    }
  }, [])
}
