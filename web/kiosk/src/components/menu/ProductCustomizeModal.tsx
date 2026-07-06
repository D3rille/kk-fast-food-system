import { useMemo, useState } from "react"
import { IconCheck, IconMinus, IconPlus } from "@tabler/icons-react"
import type { ModifierGroup, Product, SelectedModifier } from "@/types/api"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { resolveImageUrl } from "@/api/client"
import { cn } from "@/lib/utils"

interface ProductCustomizeModalProps {
  product: Product
  groups: ModifierGroup[]
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (selectedModifiers: SelectedModifier[], unitPrice: number, quantity: number) => void
}

function formatPrice(centavos: number): string {
  return `₱${(centavos / 100).toFixed(2)}`
}

// Selections are keyed by group id -> set of chosen option ids, initialised from each group's
// is_default option so a required group is always satisfiable without the customer touching it.
function buildInitialSelection(groups: ModifierGroup[]): Record<string, string[]> {
  const initial: Record<string, string[]> = {}
  for (const group of groups) {
    const defaults = group.options.filter((o) => o.is_default).map((o) => o.id)
    initial[group.id] = defaults
  }
  return initial
}

export function ProductCustomizeModal({
  product,
  groups,
  open,
  onOpenChange,
  onConfirm,
}: ProductCustomizeModalProps) {
  // The parent remounts this component (via a changing `key`) each time it opens, so this
  // initial state is fresh — defaults selected, quantity reset — on every "Add" tap.
  const [selection, setSelection] = useState<Record<string, string[]>>(() =>
    buildInitialSelection(groups),
  )
  const [quantity, setQuantity] = useState(1)

  function toggleOption(group: ModifierGroup, optionId: string) {
    setSelection((prev) => {
      const current = prev[group.id] ?? []
      if (group.max_selection === 1) {
        return { ...prev, [group.id]: [optionId] }
      }
      if (current.includes(optionId)) {
        return { ...prev, [group.id]: current.filter((id) => id !== optionId) }
      }
      if (current.length >= group.max_selection) {
        return prev
      }
      return { ...prev, [group.id]: [...current, optionId] }
    })
  }

  const { extraPrice, isValid, selectedModifiers } = useMemo(() => {
    let extra = 0
    const modifiers: SelectedModifier[] = []
    let valid = true
    for (const group of groups) {
      const chosen = selection[group.id] ?? []
      if (chosen.length < group.min_selection || chosen.length > group.max_selection) {
        valid = false
      }
      for (const optionId of chosen) {
        const option = group.options.find((o) => o.id === optionId)
        if (!option) continue
        extra += option.extra_price
        modifiers.push({
          groupId: group.id,
          groupName: group.name,
          optionId: option.id,
          name: option.name,
          extraPrice: option.extra_price,
        })
      }
    }
    return { extraPrice: extra, isValid: valid, selectedModifiers: modifiers }
  }, [groups, selection])

  const unitPrice = product.base_price + extraPrice

  function handleConfirm() {
    if (!isValid) return
    onConfirm(selectedModifiers, unitPrice, quantity)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="p-0">
        <DialogHeader className="flex-row items-center gap-4">
          <div className="h-16 w-16 shrink-0 overflow-hidden rounded-xl bg-secondary/50">
            {product.image_url ? (
              <img
                src={resolveImageUrl(product.image_url)}
                alt={product.name}
                className="h-full w-full object-cover"
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center text-2xl opacity-40">
                🍽️
              </div>
            )}
          </div>
          <div className="min-w-0 flex-1">
            <DialogTitle className="line-clamp-1">{product.name}</DialogTitle>
            <DialogDescription className="mt-0.5">
              Customize your order
            </DialogDescription>
          </div>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto px-6 py-2">
          {groups.map((group) => (
            <div key={group.id} className="py-3 border-b border-border last:border-0">
              <div className="flex items-center justify-between mb-2.5">
                <h3 className="font-semibold text-sm text-foreground">{group.name}</h3>
                <span
                  className={cn(
                    "text-[10px] font-bold uppercase tracking-wide px-2 py-0.5 rounded-full",
                    group.is_required
                      ? "bg-primary/10 text-primary"
                      : "bg-muted text-muted-foreground",
                  )}
                >
                  {group.is_required
                    ? "Required"
                    : group.max_selection > 1
                      ? `Up to ${group.max_selection}`
                      : "Optional"}
                </span>
              </div>

              <div className="flex flex-col gap-2">
                {group.options.map((option) => {
                  const checked = (selection[group.id] ?? []).includes(option.id)
                  return (
                    <button
                      key={option.id}
                      type="button"
                      onClick={() => toggleOption(group, option.id)}
                      className={cn(
                        "flex items-center justify-between rounded-xl border px-4 py-3 text-left transition-colors min-h-12",
                        checked
                          ? "border-primary bg-primary/8"
                          : "border-border bg-background hover:bg-muted/50",
                      )}
                    >
                      <span className="flex items-center gap-2.5">
                        <span
                          className={cn(
                            "flex h-5 w-5 shrink-0 items-center justify-center border-2",
                            group.max_selection === 1 ? "rounded-full" : "rounded-md",
                            checked
                              ? "border-primary bg-primary text-primary-foreground"
                              : "border-muted-foreground/40",
                          )}
                        >
                          {checked && <IconCheck size={13} stroke={3} />}
                        </span>
                        <span className="text-sm font-medium">{option.name}</span>
                      </span>
                      {option.extra_price > 0 && (
                        <span className="font-mono text-xs font-semibold text-muted-foreground tabular-nums">
                          +{formatPrice(option.extra_price)}
                        </span>
                      )}
                    </button>
                  )
                })}
              </div>
            </div>
          ))}
        </div>

        <DialogFooter>
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="icon"
                className="h-11 w-11 rounded-xl"
                onClick={() => setQuantity((q) => Math.max(1, q - 1))}
              >
                <IconMinus size={16} />
              </Button>
              <span className="w-6 text-center font-mono font-bold tabular-nums">{quantity}</span>
              <Button
                variant="outline"
                size="icon"
                className="h-11 w-11 rounded-xl"
                onClick={() => setQuantity((q) => q + 1)}
              >
                <IconPlus size={16} />
              </Button>
            </div>
            <Button
              size="lg"
              className="flex-1 min-h-12 font-bold"
              disabled={!isValid}
              onClick={handleConfirm}
            >
              Add · {formatPrice(unitPrice * quantity)}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
