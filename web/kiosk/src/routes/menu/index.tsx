import { useEffect, useState } from "react"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useQuery } from "@tanstack/react-query"
import { IconShoppingCart } from "@tabler/icons-react"
import { getCategories, getProducts } from "@/api/menu"
import { CategoryTabs } from "@/components/menu/CategoryTabs"
import { ProductCard } from "@/components/menu/ProductCard"
import { useCart } from "@/contexts/cart"
import { Button } from "@/components/ui/button"
import type { Product } from "@/types/api"

export const Route = createFileRoute("/menu/")({ component: MenuScreen })

function MenuScreen() {
  const navigate = useNavigate()
  const { dispatch, itemCount, total } = useCart()
  const [selectedCategoryId, setSelectedCategoryId] = useState<string | undefined>(undefined)

  const { data: categories = [], isLoading: loadingCategories } = useQuery({
    queryKey: ["categories"],
    queryFn: getCategories,
    select: (data) => [...data].sort((a, b) => a.sort_order - b.sort_order),
  })

  const { data: products = [], isLoading: loadingProducts } = useQuery({
    queryKey: ["products", selectedCategoryId],
    queryFn: () => getProducts(selectedCategoryId),
    enabled: !!selectedCategoryId,
  })

  useEffect(() => {
    if (!selectedCategoryId && categories.length > 0) {
      setSelectedCategoryId(categories[0].id)
    }
  }, [categories, selectedCategoryId])

  const activeCategoryId = selectedCategoryId ?? categories[0]?.id

  const visibleProducts = activeCategoryId
    ? products.filter((p) => p.category_id === activeCategoryId)
    : products

  function handleAdd(product: Product) {
    dispatch({ type: "ADD_ITEM", product })
  }

  const isLoading = loadingCategories || loadingProducts

  return (
    <div className="flex h-svh flex-col bg-background">
      <header className="flex items-center justify-between px-6 py-4 border-b border-border shrink-0 bg-card">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground text-lg">
            🍔
          </div>
          <div>
            <h1 className="font-heading font-bold text-base leading-tight">NextGen Kitchen</h1>
            <p className="text-xs text-muted-foreground leading-none mt-0.5">Pick your favorites</p>
          </div>
        </div>

        <Button
          variant="outline"
          size="lg"
          className="relative gap-2 min-h-11"
          onClick={() => navigate({ to: "/cart" })}
        >
          <IconShoppingCart size={18} />
          {itemCount > 0 ? (
            <>
              <span className="font-mono font-bold text-sm tabular-nums">{itemCount}</span>
              <span className="text-muted-foreground text-sm">·</span>
              <span className="font-mono font-bold text-sm tabular-nums text-primary">
                ₱{(total / 100).toFixed(0)}
              </span>
            </>
          ) : (
            <span className="font-semibold text-sm">Cart</span>
          )}
          {itemCount > 0 && (
            <span className="absolute -top-1.5 -right-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-primary text-[9px] font-bold text-primary-foreground">
              {itemCount > 9 ? "9+" : itemCount}
            </span>
          )}
        </Button>
      </header>

      <div className="px-6 pt-4 pb-3 shrink-0 bg-card border-b border-border">
        {isLoading ? (
          <div className="flex gap-2">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="h-11 w-24 rounded-full bg-muted animate-pulse" />
            ))}
          </div>
        ) : (
          <CategoryTabs
            categories={categories}
            selected={activeCategoryId}
            onSelect={setSelectedCategoryId}
          />
        )}
      </div>

      <main className="flex-1 overflow-y-auto px-4 py-4 sm:px-6 sm:py-5">
        {isLoading ? (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div key={i} className="rounded-2xl border border-border overflow-hidden bg-card">
                <div className="aspect-square bg-muted animate-pulse" />
                <div className="p-3 flex flex-col gap-2">
                  <div className="h-4 w-3/4 rounded bg-muted animate-pulse" />
                  <div className="h-3 w-1/2 rounded bg-muted animate-pulse" />
                  <div className="h-5 w-16 rounded bg-muted animate-pulse mt-1" />
                  <div className="h-10 rounded-xl bg-muted animate-pulse mt-1" />
                </div>
              </div>
            ))}
          </div>
        ) : visibleProducts.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <span className="text-4xl mb-3 opacity-40">🍽️</span>
            <p className="font-semibold text-foreground/60">Nothing here yet.</p>
            <p className="text-sm text-muted-foreground mt-1">Try another category.</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
            {visibleProducts.map((product) => (
              <ProductCard key={product.id} product={product} onAdd={handleAdd} />
            ))}
          </div>
        )}
      </main>

      {itemCount > 0 && (
        <div className="shrink-0 border-t border-border bg-card px-4 py-3">
          <Button
            size="lg"
            className="w-full gap-2 font-bold"
            onClick={() => navigate({ to: "/cart" })}
          >
            <IconShoppingCart size={18} />
            Review Order · {itemCount} {itemCount === 1 ? "item" : "items"}
          </Button>
        </div>
      )}
    </div>
  )
}
