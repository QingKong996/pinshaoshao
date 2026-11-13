import { LoaderCircle } from "lucide-react"
import { useTranslation } from "react-i18next"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog"
import type { ProductSKU } from "../types"
import { formatSpecsLabel } from "./format-specs"

export function DeleteSKUDialog({
  sku,
  pending,
  onOpenChange,
  onConfirm,
}: {
  sku: ProductSKU | null
  pending: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  const { t } = useTranslation("console")
  const specs = sku
    ? formatSpecsLabel(sku, t("productDetail.skus.defaultSpecs"))
    : ""

  return (
    <AlertDialog open={sku !== null} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {t("productDetail.skus.deleteDialog.title")}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {t("productDetail.skus.deleteDialog.description", { specs })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={pending}>
            {t("common.cancel")}
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={pending}
            onClick={onConfirm}
          >
            {pending ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : null}
            {t("productDetail.skus.deleteDialog.confirm")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
