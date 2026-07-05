import { IconMinus, IconPlus, IconTrash } from "@tabler/icons-react"
import type { CartItem as CartItemType } from "@/types/api"
import { Button } from "@/components/ui/button"

interface CartItemProps {
  item: CartItemType
  onIncrement: () => void
  onDecrement: () => void
  onRemove: () => void
}

function formatPrice(centavos: number): string {
  return `₱${(centavos / 100).toFixed(2)}`
}

export function CartItem({ item, onIncrement, onDecrement, onRemove }: CartItemProps) {
  return (
    <div className="flex items-center gap-3 py-4 border-b border-border last:border-0">
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-sm leading-snug line-clamp-2">{item.product.name}</p>
        <p className="text-xs text-muted-foreground font-mono mt-0.5">
          {formatPrice(item.product.base_price)} each
        </p>
      </div>

      <div className="flex items-center gap-1.5 shrink-0">
        <Button
          variant="outline"
          size="icon"
          onClick={onDecrement}
          className="h-10 w-10 rounded-xl"
        >
          <IconMinus size={15} />
        </Button>
        <span className="w-7 text-center font-mono font-bold text-sm tabular-nums">
          {item.quantity}
        </span>
        <Button
          variant="outline"
          size="icon"
          onClick={onIncrement}
          className="h-10 w-10 rounded-xl"
        >
          <IconPlus size={15} />
        </Button>
      </div>

      <div className="text-right shrink-0 min-w-[68px]">
        <p className="font-mono font-bold text-sm text-primary">
          {formatPrice(item.product.base_price * item.quantity)}
        </p>
      </div>

      <Button
        variant="ghost"
        size="icon"
        onClick={onRemove}
        className="text-muted-foreground/60 hover:text-destructive hover:bg-destructive/10 h-10 w-10 shrink-0"
      >
        <IconTrash size={16} />
      </Button>
    </div>
  )
}
