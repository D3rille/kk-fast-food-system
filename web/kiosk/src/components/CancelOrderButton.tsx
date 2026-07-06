import { useState } from "react"
import { useNavigate } from "@tanstack/react-router"
import { IconX } from "@tabler/icons-react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { useCart } from "@/contexts/cart"

// Cancels the customer's turn before any order has been created server-side (used on /menu and
// /payment, prior to "Place Order"). This only clears the local cart and returns to the start
// screen — it never calls the backend, since there is no order yet to cancel.
export function CancelOrderButton({ disabled = false }: { disabled?: boolean }) {
  const navigate = useNavigate()
  const { dispatch } = useCart()
  const [open, setOpen] = useState(false)

  function handleConfirm() {
    dispatch({ type: "CLEAR_CART" })
    setOpen(false)
    navigate({ to: "/" })
  }

  return (
    <>
      <Button
        variant="ghost"
        size="lg"
        className="gap-1.5 text-muted-foreground hover:text-destructive"
        onClick={() => setOpen(true)}
        disabled={disabled}
      >
        <IconX size={18} />
        <span className="hidden sm:inline">Cancel</span>
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Cancel your order?</DialogTitle>
            <DialogDescription>
              Your cart will be cleared and you&apos;ll return to the start screen.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="flex-row gap-2">
            <Button variant="outline" className="flex-1" onClick={() => setOpen(false)}>
              Keep Ordering
            </Button>
            <Button variant="destructive" className="flex-1" onClick={handleConfirm}>
              Yes, Cancel
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
