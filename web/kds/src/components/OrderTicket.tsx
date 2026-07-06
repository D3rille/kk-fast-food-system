import type { OrderDetail } from "@/types/api"
import { Button } from "@/components/ui/button"

interface OrderTicketProps {
  order: OrderDetail
  now: number
  onStartPreparation: (id: string) => void
  onMarkReady: (id: string) => void
  onComplete: (id: string) => void
  isAdvancing: boolean
}

const STATUS_LABELS: Record<string, string> = {
  paid: "Queued",
  in_preparation: "In Prep",
  ready_for_pickup: "Ready",
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

export function OrderTicket({
  order,
  now,
  onStartPreparation,
  onMarkReady,
  onComplete,
  isAdvancing,
}: OrderTicketProps) {
  const elapsed = getElapsedSeconds(order.created_at, now)
  const agingColor = getAgingColor(elapsed)
  const srcLabel = SOURCE_LABELS[order.source] ?? order.source.toUpperCase()
  const srcStyle = SOURCE_STYLE[order.source] ?? {
    bg: "rgba(107,114,128,0.15)",
    color: "#9ca3af",
    border: "rgba(107,114,128,0.35)",
  }
  const totalPesos = (order.total_amount / 100).toFixed(2)
  const statusLabel = STATUS_LABELS[order.status] ?? order.status

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

      {/* Items + modifiers — modifiers are bolded so the kitchen never misses a customization */}
      {order.items.length > 0 && (
        <div className="flex flex-col divide-y divide-border/60 px-4">
          {order.items.map((item) => (
            <div key={item.id} className="py-2">
              <p className="text-sm font-semibold text-foreground">
                {item.quantity}× {item.product_name}
              </p>
              {item.modifiers.length > 0 && (
                <p className="mt-0.5 text-xs font-bold uppercase tracking-wide" style={{ color: "#ef4444" }}>
                  {item.modifiers.map((m) => m.name).join(" · ")}
                </p>
              )}
            </div>
          ))}
        </div>
      )}

      <div className="border-t border-border" />

      {/* Total + status label */}
      <div className="flex items-center justify-between px-4 py-3">
        <span className="text-lg font-semibold text-foreground">₱{totalPesos}</span>
        <span
          className="text-xs font-bold uppercase tracking-wide"
          style={{
            color:
              order.status === "paid"
                ? "#fbbf24"
                : order.status === "in_preparation"
                  ? "#60a5fa"
                  : "#c084fc",
          }}
        >
          {statusLabel}
        </span>
      </div>

      {/* Action button */}
      <div className="px-3 pb-3">
        {order.status === "paid" && (
          <Button
            className="w-full font-bold h-11 text-white"
            style={{ backgroundColor: "#2563eb" }}
            onClick={() => onStartPreparation(order.id)}
            disabled={isAdvancing}
          >
            Start Prep
          </Button>
        )}
        {order.status === "in_preparation" && (
          <Button
            className="w-full font-bold h-11 text-white"
            style={{ backgroundColor: "#16a34a" }}
            onClick={() => onMarkReady(order.id)}
            disabled={isAdvancing}
          >
            Ready for Pickup
          </Button>
        )}
        {order.status === "ready_for_pickup" && (
          <Button
            className="w-full font-bold h-11 text-white"
            style={{ backgroundColor: "#9333ea" }}
            onClick={() => onComplete(order.id)}
            disabled={isAdvancing}
          >
            Complete / Picked Up
          </Button>
        )}
      </div>
    </div>
  )
}
