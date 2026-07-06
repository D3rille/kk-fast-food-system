interface DisplayCardProps {
  orderNumber: number
}

export function DisplayCard({ orderNumber }: DisplayCardProps) {
  return (
    <div className="flex items-center justify-center rounded-lg bg-card border border-border py-6">
      <span className="text-6xl font-black tabular-nums leading-none text-foreground">
        #{orderNumber}
      </span>
    </div>
  )
}
