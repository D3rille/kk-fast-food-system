import { useState } from "react"
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  createColumnHelper,
  type SortingState,
} from "@tanstack/react-table"
import { useQuery } from "@tanstack/react-query"
import { ChevronUp, ChevronDown, ArrowUpDown } from "lucide-react"
import { fetchCategories } from "@/api/categories"
import type { Category } from "@/types/api"

const columnHelper = createColumnHelper<Category>()

const columns = [
  columnHelper.accessor("name", {
    header: "Category",
    cell: (info) => <span className="font-semibold text-sm">{info.getValue()}</span>,
  }),
  columnHelper.accessor("sort_order", {
    header: "Sort Order",
    cell: (info) => (
      <span className="text-sm font-mono text-muted-foreground">{info.getValue()}</span>
    ),
  }),
  columnHelper.accessor("is_active", {
    header: "Status",
    enableSorting: false,
    cell: (info) => {
      const active = info.getValue()
      return (
        <span
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 5,
            padding: "3px 10px",
            borderRadius: 9999,
            fontSize: 12,
            fontWeight: 600,
            background: active ? "#10B98118" : "#F8717118",
            color: active ? "#10B981" : "#F87171",
            border: `1px solid ${active ? "#10B98130" : "#F8717130"}`,
          }}
        >
          {active ? "Active" : "Inactive"}
        </span>
      )
    },
  }),
]

export default function CategoriesPage() {
  const [sorting, setSorting] = useState<SortingState>([{ id: "sort_order", desc: false }])

  const { data: categories = [], isLoading, isError } = useQuery({
    queryKey: ["admin-categories"],
    queryFn: fetchCategories,
  })

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data: categories,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  return (
    <div className="p-6 space-y-5">
      <div>
        <h1 className="text-xl font-bold">Categories</h1>
        <p className="text-sm text-muted-foreground mt-0.5">
          {isLoading ? "Loading…" : `${categories.length} categories`}
        </p>
      </div>

      {isError && (
        <div className="rounded-lg bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">
          Failed to load categories. Check your connection and try again.
        </div>
      )}

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
                <td colSpan={3} className="px-4 py-16 text-center text-sm text-muted-foreground">
                  Loading categories…
                </td>
              </tr>
            ) : table.getRowModel().rows.length === 0 ? (
              <tr>
                <td colSpan={3} className="px-4 py-16 text-center text-sm text-muted-foreground">
                  No categories found.
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
    </div>
  )
}
