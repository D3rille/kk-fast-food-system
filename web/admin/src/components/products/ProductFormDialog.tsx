import { useEffect, useMemo, useState } from "react"
import { useForm } from "@tanstack/react-form"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { ImageIcon, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { createProduct, updateProduct } from "@/api/products"
import { resolveImageUrl } from "@/api/client"
import type { Category, Product } from "@/types/api"

interface ProductFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  categories: Category[]
  product?: Product
}

const inputClass =
  "w-full rounded-lg border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-shadow"

export function ProductFormDialog({ open, onOpenChange, categories, product }: ProductFormDialogProps) {
  const queryClient = useQueryClient()
  const isEdit = Boolean(product)

  const [imageFile, setImageFile] = useState<File | null>(null)
  const [removeImage, setRemoveImage] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Dialog content unmounts when closed (Radix Presence), which naturally resets this
  // local state — no reset-on-open effect needed.
  const imagePreview = useMemo(() => (imageFile ? URL.createObjectURL(imageFile) : null), [imageFile])
  useEffect(() => {
    return () => {
      if (imagePreview) URL.revokeObjectURL(imagePreview)
    }
  }, [imagePreview])

  const form = useForm({
    defaultValues: {
      categoryId: product?.category_id ?? "",
      name: product?.name ?? "",
      description: product?.description ?? "",
      basePrice: product ? (product.base_price / 100).toFixed(2) : "",
      isAvailable: product?.is_available ?? true,
    },
    onSubmit: async ({ value }) => {
      setError(null)

      const centavos = Math.round(parseFloat(value.basePrice || "0") * 100)
      if (!value.categoryId || !value.name || !Number.isFinite(centavos) || centavos <= 0) {
        setError("Category, name, and a valid price are required.")
        return
      }

      const formData = new FormData()
      formData.append("category_id", value.categoryId)
      formData.append("name", value.name)
      formData.append("description", value.description)
      formData.append("base_price", String(centavos))
      if (isEdit) {
        formData.append("is_available", String(value.isAvailable))
      }
      if (imageFile) {
        formData.append("image", imageFile)
      } else if (isEdit && removeImage) {
        formData.append("remove_image", "true")
      }

      try {
        await mutation.mutateAsync(formData)
        onOpenChange(false)
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to save menu item.")
      }
    },
  })

  const mutation = useMutation({
    mutationFn: (formData: FormData) =>
      isEdit && product ? updateProduct(product.id, formData) : createProduct(formData),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin-products"] })
    },
  })

  const currentImageUrl =
    product?.image_url && !removeImage ? (resolveImageUrl(product.image_url) ?? null) : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Menu Item" : "New Menu Item"}</DialogTitle>
        </DialogHeader>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            e.stopPropagation()
            void form.handleSubmit()
          }}
          className="space-y-4"
        >
          <div className="grid grid-cols-2 gap-3">
            <form.Field name="categoryId">
              {(field) => (
                <div className="space-y-1.5 col-span-2 sm:col-span-1">
                  <label htmlFor="categoryId" className="text-sm font-medium">
                    Category
                  </label>
                  <select
                    id="categoryId"
                    className={inputClass}
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                  >
                    <option value="" disabled>
                      Select a category
                    </option>
                    {categories.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </form.Field>

            <form.Field name="basePrice">
              {(field) => (
                <div className="space-y-1.5 col-span-2 sm:col-span-1">
                  <label htmlFor="basePrice" className="text-sm font-medium">
                    Price (₱)
                  </label>
                  <input
                    id="basePrice"
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    className={inputClass}
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            </form.Field>
          </div>

          <form.Field name="name">
            {(field) => (
              <div className="space-y-1.5">
                <label htmlFor="name" className="text-sm font-medium">
                  Name
                </label>
                <input
                  id="name"
                  autoFocus
                  className={inputClass}
                  placeholder="e.g. Classic Cheeseburger"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              </div>
            )}
          </form.Field>

          <form.Field name="description">
            {(field) => (
              <div className="space-y-1.5">
                <label htmlFor="description" className="text-sm font-medium">
                  Description
                </label>
                <textarea
                  id="description"
                  rows={2}
                  className={inputClass}
                  placeholder="Optional short description"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              </div>
            )}
          </form.Field>

          {isEdit && (
            <form.Field name="isAvailable">
              {(field) => (
                <label className="flex items-center gap-2 text-sm font-medium">
                  <input
                    type="checkbox"
                    checked={field.state.value}
                    onChange={(e) => field.handleChange(e.target.checked)}
                  />
                  Available for sale
                </label>
              )}
            </form.Field>
          )}

          <div className="space-y-1.5">
            <span className="text-sm font-medium">Image</span>
            <div className="flex items-center gap-3">
              <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-lg border bg-secondary/50 flex items-center justify-center">
                {imagePreview || currentImageUrl ? (
                  <img
                    src={imagePreview ?? currentImageUrl ?? undefined}
                    alt="Preview"
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <ImageIcon size={20} className="opacity-40" />
                )}
              </div>
              <div className="flex-1 space-y-1.5">
                <input
                  type="file"
                  accept="image/jpeg,image/png,image/webp,image/gif"
                  onChange={(e) => {
                    setImageFile(e.target.files?.[0] ?? null)
                    setRemoveImage(false)
                  }}
                  className="w-full text-xs file:mr-2 file:rounded-md file:border-0 file:bg-secondary file:px-2.5 file:py-1.5 file:text-xs file:font-medium hover:file:bg-secondary/70"
                />
                {isEdit && currentImageUrl && !imageFile && (
                  <button
                    type="button"
                    onClick={() => setRemoveImage(true)}
                    className="inline-flex items-center gap-1 text-xs text-destructive hover:underline"
                  >
                    <X size={12} />
                    Remove current image
                  </button>
                )}
              </div>
            </div>
          </div>

          {error && (
            <p className="text-sm text-destructive bg-destructive/10 rounded-lg px-3 py-2">
              {error}
            </p>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <form.Subscribe selector={(s) => s.isSubmitting}>
              {(isSubmitting) => (
                <Button type="submit" disabled={isSubmitting || mutation.isPending}>
                  {isSubmitting || mutation.isPending ? "Saving…" : isEdit ? "Save Changes" : "Create Item"}
                </Button>
              )}
            </form.Subscribe>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
