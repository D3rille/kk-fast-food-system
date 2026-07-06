import { IconLoader2, IconPlus } from "@tabler/icons-react"
import type { Product } from "@/types/api"
import { Button } from "@/components/ui/button"
import { resolveImageUrl } from "@/api/client"

interface ProductCardProps {
  product: Product
  onAdd: (product: Product) => void
  isLoading?: boolean
}

function formatPrice(centavos: number): string {
  return `₱${(centavos / 100).toFixed(2)}`
}

export function ProductCard({ product, onAdd, isLoading }: ProductCardProps) {
  return (
    <div className="bg-card rounded-2xl border border-border flex flex-col overflow-hidden shadow-sm transition-shadow hover:shadow-md active:scale-[0.98] transition-transform">
      <div className="relative aspect-square bg-secondary/50 w-full shrink-0 overflow-hidden">
        {product.image_url ? (
          <img
            src={resolveImageUrl(product.image_url)}
            alt={product.name}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="h-full w-full flex items-center justify-center">
            <span className="text-5xl opacity-40">🍽️</span>
          </div>
        )}
        {!product.is_available && (
          <div className="absolute inset-0 bg-background/70 flex items-center justify-center">
            <span className="text-xs font-semibold text-muted-foreground bg-background/90 px-3 py-1 rounded-full">
              Unavailable
            </span>
          </div>
        )}
      </div>

      <div className="flex flex-col px-3 pt-3 pb-1 flex-1 gap-1">
        <h3 className="font-semibold text-sm leading-tight line-clamp-2 text-foreground">
          {product.name}
        </h3>
        {product.description && (
          <p className="text-xs text-muted-foreground line-clamp-1 leading-snug">
            {product.description}
          </p>
        )}
        <p className="font-mono font-bold text-lg text-primary mt-auto pt-2">
          {formatPrice(product.base_price)}
        </p>
      </div>

      <div className="px-3 pb-3 pt-1">
        <Button
          size="lg"
          onClick={() => onAdd(product)}
          disabled={!product.is_available || isLoading}
          className="w-full gap-1.5"
        >
          {isLoading ? <IconLoader2 size={16} className="animate-spin" /> : <IconPlus size={16} />}
          Add
        </Button>
      </div>
    </div>
  )
}
