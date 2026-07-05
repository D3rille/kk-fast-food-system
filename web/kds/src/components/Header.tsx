import { useState, useEffect } from "react"

interface HeaderProps {
  connected: boolean
  pending: number
  inPrep: number
}

export function Header({ connected, pending, inPrep }: HeaderProps) {
  const [time, setTime] = useState(new Date())

  useEffect(() => {
    const id = setInterval(() => setTime(new Date()), 1000)
    return () => clearInterval(id)
  }, [])

  return (
    <header className="flex items-center justify-between px-6 py-3 border-b border-border bg-card shrink-0">
      <div className="flex items-center gap-2">
        <span className="text-xs font-bold tracking-widest uppercase text-muted-foreground">
          Kitchen
        </span>
        <span className="text-xl font-bold text-foreground">Display</span>
      </div>

      <div className="flex items-center gap-6">
        <div className="flex items-center gap-4 text-sm">
          <span className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-amber-400" />
            <span className="text-muted-foreground">
              {pending} {pending === 1 ? "queued" : "queued"}
            </span>
          </span>
          <span className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-blue-400" />
            <span className="text-muted-foreground">
              {inPrep} in prep
            </span>
          </span>
        </div>

        <span
          className="flex items-center gap-1.5 text-xs font-medium"
          style={{ color: connected ? "#34d399" : "#f87171" }}
        >
          <span
            className="size-2 rounded-full"
            style={{ backgroundColor: connected ? "#34d399" : "#f87171" }}
          />
          {connected ? "Connected" : "Reconnecting…"}
        </span>

        <time className="font-mono text-sm tabular-nums text-foreground">
          {time.toLocaleTimeString()}
        </time>
      </div>
    </header>
  )
}
