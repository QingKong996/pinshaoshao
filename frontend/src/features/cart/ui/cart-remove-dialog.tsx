import { LoaderCircle } from "lucide-react"
import { useTranslation } from "react-i18next"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog"
import { Button } from "@/shared/ui/button"
import type { CartItem } from "../types"

export function CartRemoveDialog({
  item,
  isPending,
  onClose,
  onConfirm,
}: {
  item: CartItem | null
  isPending: boolean
  onClose: () => void
  onConfirm: () => void
}) {
  const { t } = useTranslation("cart")

  return (
    <AlertDialog
      open={item !== null}
      onOpenChange={(open) => {
        if (!open && !isPending) {
          onClose()
        }
      }}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {t("removeDialog.title")}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {t("removeDialog.description", {
              name: item?.product_name ?? "",
            })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>
            {t("removeDialog.cancel")}
          </AlertDialogCancel>
          <Button
            variant="destructive"
            disabled={isPending}
            onClick={onConfirm}
          >
            {isPending ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : null}
            {t("removeDialog.confirm")}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
