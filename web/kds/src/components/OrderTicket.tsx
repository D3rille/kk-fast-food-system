import type { Order, OrderStatus } from "@/types/api"
import { Button } from "@/components/ui/button"

interface OrderTicketProps {
  order: Order
  now: number
  onAdvance: (id: string, status: OrderStatus) => void
  isAdvancing: boolean
}

function getElapsedSeconds(createdAt: string, now: number): number {
  return Math.floor((now - new Date(createdAt).getTime()) / 1000)
}

function formatElapsed(seconds: number): string {
  const m = Math.floor(Math.abs(seconds) / 60)
  const s = Math.abs(seconds) % 60
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`
}

function getAgingColor(seconds: number): string {
  if (seconds < 300) return "#22c55e"  // green — fresh (< 5 min)
  if (seconds < 600) return "#f59e0b"  // amber — warn (5–10 min)
  return "#ef4444"                      // red — urgent (> 10 min)
}

const SOURCE_LABELS: Record<string, string> = {
  kiosk: "KIOSK",
  cashier: "CASHIER",
  mobile: "MOBILE",
  online: "ONLINE",
}

const SOURCE_STYLE: Record<string, { bg: string; color: string; border: string }> = {
  kiosk: { bg: "rgba(139,92,246,0.15)", color: "#c4b5fd", border: "rgba(139,92,246,0.35)" },
  cashier: { bg: "rgba(59,130,246,0.15)", color: "#93c5fd", border: "rgba(59,130,246,0.35)" },
  mobile: { bg: "rgba(20,184,166,0.15)", color: "#5eead4", border: "rgba(20,184,166,0.35)" },
  online: { bg: "rgba(249,115,22,0.15)", color: "#fdba74", border: "rgba(249,115,22,0.35)" },
}

export function OrderTicket({ order, now, onAdvance, isAdvancing }: OrderTicketProps) {
  const elapsed = getElapsedSeconds(order.created_at, now)
  const agingColor = getAgingColor(elapsed)
  const srcLabel = SOURCE_LABELS[order.source] ?? order.source.toUpperCase()
  const srcStyle = SOURCE_STYLE[order.source] ?? {
    bg: "rgba(107,114,128,0.15)",
    color: "#9ca3af",
    border: "rgba(107,114,128,0.35)",
  }
  const totalPesos = (order.total_amount / 100).toFixed(2)
  const isQueued = order.status === "paid"

  return (
    <div
      className="flex flex-col rounded-lg bg-card border border-border overflow-hidden"
      style={{ borderLeftColor: agingColor, borderLeftWidth: "4px" }}
    >
      {/* Order number + source badge */}
      <div className="flex items-center justify-between px-4 pt-4 pb-1">
        <span className="text-5xl font-black tabular-nums leading-none text-foreground">
          #{order.order_number}
        </span>
        <span
          className="text-xs font-bold px-2 py-0.5 rounded"
          style={{
            backgroundColor: srcStyle.bg,
            color: srcStyle.color,
            border: `1px solid ${srcStyle.border}`,
          }}
        >
          {srcLabel}
        </span>
      </div>

      {/* Elapsed timer — the visual signature of this KDS */}
      <div className="px-4 pb-4">
        <span
          className="text-3xl font-bold tabular-nums"
          style={{ color: agingColor, fontVariantNumeric: "tabular-nums" }}
        >
          {formatElapsed(elapsed)}
        </span>
      </div>

      <div className="border-t border-border" />

      {/* Total + status label */}
      <div className="flex items-center justify-between px-4 py-3">
        <span className="text-lg font-semibold text-foreground">₱{totalPesos}</span>
        <span
          className="text-xs font-bold uppercase tracking-wide"
          style={{ color: isQueued ? "#fbbf24" : "#60a5fa" }}
        >
          {isQueued ? "Queued" : "In Prep"}
        </span>
      </div>

      {/* Action button */}
      <div className="px-3 pb-3">
        {isQueued ? (
          <Button
            className="w-full font-bold h-11 text-white"
            style={{ backgroundColor: "#2563eb" }}
            onClick={() => onAdvance(order.id, "in_preparation")}
            disabled={isAdvancing}
          >
            Start Prep
          </Button>
        ) : (
          <Button
            className="w-full font-bold h-11 text-white"
            style={{ backgroundColor: "#16a34a" }}
            onClick={() => onAdvance(order.id, "ready_for_pickup")}
            disabled={isAdvancing}
          >
            Ready for Pickup
          </Button>
        )}
      </div>
    </div>
  )
}
