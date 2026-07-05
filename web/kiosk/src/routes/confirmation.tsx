import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { IconCircleCheckFilled } from "@tabler/icons-react"
import { Button } from "@/components/ui/button"

export const Route = createFileRoute("/confirmation")({
  validateSearch: (search: Record<string, unknown>) => ({
    orderNumber: Number(search.orderNumber ?? 0),
  }),
  component: ConfirmationScreen,
})

function ConfirmationScreen() {
  const navigate = useNavigate()
  const { orderNumber } = Route.useSearch()

  return (
    <div
      className="flex min-h-svh flex-col items-center justify-center px-8 text-center"
      style={{ background: "#0d0906" }}
    >
      <div className="flex flex-col items-center gap-5 max-w-sm w-full">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/15 border border-primary/30">
          <IconCircleCheckFilled size={36} className="text-primary" />
        </div>

        <p className="font-mono text-xs tracking-[0.2em] uppercase text-white/40">
          Order Confirmed
        </p>

        <div className="leading-none">
          <span
            className="font-mono font-black text-primary"
            style={{ fontSize: "clamp(80px, 20vw, 120px)", lineHeight: 1 }}
          >
            #{orderNumber > 0 ? orderNumber : "—"}
          </span>
        </div>

        <div className="rounded-2xl border border-white/10 bg-white/5 p-5 w-full mt-1">
          <p className="font-semibold text-base text-white">Order placed!</p>
          <p className="text-sm text-white/55 mt-1 leading-relaxed">
            Please proceed to the cashier to pay. Your number will be called when your order is ready.
          </p>
        </div>

        <div className="w-full mt-3">
          <Button
            size="lg"
            className="w-full min-h-14 text-base font-bold"
            onClick={() => navigate({ to: "/" })}
          >
            Start New Order
          </Button>
        </div>
      </div>
    </div>
  )
}
