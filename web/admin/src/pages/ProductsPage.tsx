import { useState, useMemo, useCallback } from "react"
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  flexRender,
  createColumnHelper,
  type SortingState,
} from "@tanstack/react-table"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { Search, ChevronUp, ChevronDown, ArrowUpDown, ImageIcon, Pencil, Trash2, Plus } from "lucide-react"
import { fetchProducts, toggleAvailability } from "@/api/products"
import { fetchCategories } from "@/api/categories"
import { resolveImageUrl } from "@/api/client"
import { Button } from "@/components/ui/button"
import { ProductFormDialog } from "@/components/products/ProductFormDialog"
import { DeleteProductDialog } from "@/components/products/DeleteProductDialog"
import type { Product, Category } from "@/types/api"

type ProductRow = Product & { categoryName: string }

const columnHelper = createColumnHelper<ProductRow>()

export default function ProductsPage() {
  const queryClient = useQueryClient()
  const [sorting, setSorting] = useState<SortingState>([{ id: "name", desc: false }])
  const [globalFilter, setGlobalFilter] = useState("")
  const [formOpen, setFormOpen] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | undefined>(undefined)
  const [deletingProduct, setDeletingProduct] = useState<Product | undefined>(undefined)

  const openCreate = useCallback(() => {
    setEditingProduct(undefined)
    setFormOpen(true)
  }, [])

  const openEdit = useCallback((product: Product) => {
    setEditingProduct(product)
    setFormOpen(true)
  }, [])

  const { data: products = [], isLoading, isError } = useQuery({
    queryKey: ["admin-products"],
    queryFn: fetchProducts,
  })

  const { data: categories = [] } = useQuery({
    queryKey: ["admin-categories"],
    queryFn: fetchCategories,
  })

  const categoryMap = useMemo(() => {
    const map: Record<string, string> = {}
    categories.forEach((c: Category) => {
      map[c.id] = c.name
    })
    return map
  }, [categories])

  const rows: ProductRow[] = useMemo(
    () => products.map((p) => ({ ...p, categoryName: categoryMap[p.category_id] ?? "—" })),
    [products, categoryMap],
  )

  const { mutate: doToggle, isPending: togglePending } = useMutation({
    mutationFn: toggleAvailability,
    onMutate: async (id: string) => {
      await queryClient.cancelQueries({ queryKey: ["admin-products"] })
      const prev = queryClient.getQueryData<Product[]>(["admin-products"])
      queryClient.setQueryData<Product[]>(["admin-products"], (old = []) =>
        old.map((p) => (p.id === id ? { ...p, is_available: !p.is_available } : p)),
      )
      return { prev }
    },
    onError: (_err, _id, ctx) => {
      queryClient.setQueryData(["admin-products"], ctx?.prev)
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin-products"] })
    },
  })

  const handleToggle = useCallback(
    (id: string) => {
      if (togglePending) return
      doToggle(id)
    },
    [doToggle, togglePending],
  )

  const columns = useMemo(
    () => [
      columnHelper.accessor("image_url", {
        header: "",
        enableSorting: false,
        cell: (info) => {
          const url = resolveImageUrl(info.getValue())
          return (
            <div className="h-10 w-10 shrink-0 overflow-hidden rounded-lg border bg-secondary/50 flex items-center justify-center">
              {url ? (
                <img src={url} alt="" className="h-full w-full object-cover" />
              ) : (
                <ImageIcon size={16} className="opacity-40" />
              )}
            </div>
          )
        },
      }),
      columnHelper.accessor("name", {
        header: "Product",
        cell: (info) => (
          <div>
            <div className="font-semibold text-sm">{info.getValue()}</div>
            {info.row.original.description && (
              <div className="text-xs text-muted-foreground mt-0.5 max-w-xs truncate">
                {info.row.original.description}
              </div>
            )}
          </div>
        ),
      }),
      columnHelper.accessor("categoryName", {
        header: "Category",
        cell: (info) => <span className="text-sm text-muted-foreground">{info.getValue()}</span>,
      }),
      columnHelper.accessor("base_price", {
        header: "Price",
        cell: (info) => (
          <span className="text-sm font-mono">₱{(info.getValue() / 100).toFixed(2)}</span>
        ),
      }),
      columnHelper.accessor("is_available", {
        header: "Status",
        enableSorting: false,
        cell: (info) => {
          const available = info.getValue()
          const id = info.row.original.id
          return (
            <button
              onClick={() => handleToggle(id)}
              disabled={togglePending}
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 6,
                padding: "4px 12px",
                borderRadius: 9999,
                fontSize: 12,
                fontWeight: 600,
                letterSpacing: "0.02em",
                cursor: togglePending ? "not-allowed" : "pointer",
                border: "none",
                transition: "all 150ms ease",
                background: available ? "#10B981" : "#3F1418",
                color: available ? "#ffffff" : "#F87171",
                opacity: togglePending ? 0.6 : 1,
              }}
            >
              <span
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: "50%",
                  background: available ? "#ffffff" : "#F87171",
                  display: "inline-block",
                  flexShrink: 0,
                }}
              />
              {available ? "Available" : "Out of Stock"}
            </button>
          )
        },
      }),
      columnHelper.display({
        id: "actions",
        header: "",
        cell: (info) => (
          <div className="flex items-center justify-end gap-1">
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => openEdit(info.row.original)}
              aria-label="Edit item"
            >
              <Pencil size={14} />
            </Button>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => setDeletingProduct(info.row.original)}
              aria-label="Delete item"
            >
              <Trash2 size={14} />
            </Button>
          </div>
        ),
      }),
    ],
    [handleToggle, togglePending, openEdit],
  )

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data: rows,
    columns,
    state: { sorting, globalFilter },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  })

  const availableCount = products.filter((p) => p.is_available).length
  const totalCount = products.length

  return (
    <div className="p-6 space-y-5">
      {/* Page header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold">Menu Items</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {isLoading ? "Loading…" : `${availableCount} of ${totalCount} items available`}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />
            <input
              className="pl-9 pr-3 py-2 text-sm rounded-lg border bg-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent w-52 transition-shadow"
              placeholder="Search items…"
              value={globalFilter}
              onChange={(e) => setGlobalFilter(e.target.value)}
            />
          </div>
          <Button onClick={openCreate} className="gap-1.5">
            <Plus size={16} />
            New Item
          </Button>
        </div>
      </div>

      {isError && (
        <div className="rounded-lg bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">
          Failed to load menu items. Check your connection and try again.
        </div>
      )}

      {/* Table */}
      <div className="rounded-xl border overflow-hidden">
        <table className="w-full">
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id} className="border-b bg-muted/30">
                {headerGroup.headers.map((header) => (
                  <th
                    key={header.id}
                    className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                  >
                    {header.isPlaceholder ? null : (
                      <div
                        role={header.column.getCanSort() ? "button" : undefined}
                        tabIndex={header.column.getCanSort() ? 0 : undefined}
                        className={
                          header.column.getCanSort()
                            ? "inline-flex items-center gap-1 cursor-pointer hover:text-foreground transition-colors select-none"
                            : "inline-flex items-center gap-1"
                        }
                        onClick={header.column.getToggleSortingHandler()}
                        onKeyDown={(e) => {
                          if (e.key === "Enter" || e.key === " ") {
                            e.preventDefault()
                            header.column.getToggleSortingHandler()?.(e)
                          }
                        }}
                      >
                        {flexRender(header.column.columnDef.header, header.getContext())}
                        {header.column.getCanSort() &&
                          (header.column.getIsSorted() === "asc" ? (
                            <ChevronUp size={12} />
                          ) : header.column.getIsSorted() === "desc" ? (
                            <ChevronDown size={12} />
                          ) : (
                            <ArrowUpDown size={12} className="opacity-40" />
                          ))}
                      </div>
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={columns.length} className="px-4 py-16 text-center text-sm text-muted-foreground">
                  Loading menu items…
                </td>
              </tr>
            ) : table.getRowModel().rows.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="px-4 py-16 text-center text-sm text-muted-foreground">
                  {globalFilter ? `No items match "${globalFilter}"` : "No menu items found."}
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row, i) => (
                <tr
                  key={row.id}
                  className={`border-b last:border-0 hover:bg-muted/20 transition-colors ${
                    i % 2 !== 0 ? "bg-muted/10" : ""
                  }`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-4 py-3">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {!isLoading && totalCount > 0 && (
        <p className="text-xs text-muted-foreground text-right">
          {table.getFilteredRowModel().rows.length} of {totalCount} items shown
        </p>
      )}

      <ProductFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        categories={categories}
        product={editingProduct}
      />
      <DeleteProductDialog
        open={Boolean(deletingProduct)}
        onOpenChange={(next) => {
          if (!next) setDeletingProduct(undefined)
        }}
        product={deletingProduct}
      />
    </div>
  )
}
