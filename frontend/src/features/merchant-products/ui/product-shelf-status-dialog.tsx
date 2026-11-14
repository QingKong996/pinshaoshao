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
import type { Product } from "../types"

export function ProductShelfStatusDialog({
  product,
  isPending,
  onClose,
  onConfirm,
}: {
  product: Product | null
  isPending: boolean
  onClose: () => void
  onConfirm: () => void
}) {
  const { t } = useTranslation("console")
  const isTakingOffShelf = product?.status === "on_shelf"
  const action = isTakingOffShelf ? "takeOff" : "putOn"

  return (
    <AlertDialog
      open={product !== null}
      onOpenChange={(open) => {
        if (!open && !isPending) {
          onClose()
        }
      }}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {t(`products.shelfStatusDialog.${action}.title`, {
              name: product?.name ?? "",
            })}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {t(`products.shelfStatusDialog.${action}.description`)}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>
            {t("common.cancel")}
          </AlertDialogCancel>
          <Button disabled={isPending} onClick={onConfirm}>
            {isPending ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : null}
            {t(`products.shelfStatusDialog.${action}.confirm`)}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
