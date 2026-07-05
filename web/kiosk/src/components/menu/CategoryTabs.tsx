import type { Category } from "@/types/api"
import { cn } from "@/lib/utils"

interface CategoryTabsProps {
  categories: Category[]
  selected: string | undefined
  onSelect: (id: string) => void
}

export function CategoryTabs({ categories, selected, onSelect }: CategoryTabsProps) {
  return (
    <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-none">
      {categories.map((cat) => (
        <button
          key={cat.id}
          onClick={() => onSelect(cat.id)}
          className={cn(
            "shrink-0 rounded-full px-5 py-2.5 text-sm font-semibold transition-colors min-h-11",
            selected === cat.id
              ? "bg-primary text-primary-foreground"
              : "bg-secondary text-secondary-foreground hover:bg-secondary/70",
          )}
        >
          {cat.name}
        </button>
      ))}
    </div>
  )
}
