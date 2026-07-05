import { useEffect } from "react"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { kioskLogin } from "@/api/auth"

export const Route = createFileRoute("/")({ component: WelcomeScreen })

function WelcomeScreen() {
  const navigate = useNavigate()

  useEffect(() => {
    kioskLogin().catch(() => {
      // credentials not configured — kiosk still works for menu browsing
    })
  }, [])

  return (
    <div
      className="flex min-h-svh flex-col items-center justify-center cursor-pointer select-none"
      style={{ background: "#0d0906" }}
      onClick={() => navigate({ to: "/menu" })}
    >
      <div className="flex flex-col items-center gap-6 text-center px-8 max-w-lg">
        <div className="flex flex-col items-center leading-none">
          <span
            className="font-heading font-black tracking-tighter text-white"
            style={{ fontSize: "clamp(64px, 15vw, 100px)", lineHeight: 1 }}
          >
            NEXTGEN
          </span>
          <span
            className="font-heading font-light text-primary tracking-[0.25em] uppercase"
            style={{ fontSize: "clamp(28px, 6vw, 44px)", lineHeight: 1.2 }}
          >
            KITCHEN
          </span>
        </div>

        <div className="h-px w-20 bg-primary/50" />

        <p className="font-mono text-xs tracking-[0.2em] uppercase text-white/35">
          Self-Service Kiosk
        </p>

        <div className="mt-6 animate-pulse">
          <div className="rounded-full bg-primary px-14 py-5 shadow-lg">
            <span className="font-heading font-bold text-xl tracking-wide text-primary-foreground">
              Tap to Order
            </span>
          </div>
        </div>

        <p className="text-xs text-white/20 mt-2">Touch anywhere on screen to begin</p>
      </div>
    </div>
  )
}
