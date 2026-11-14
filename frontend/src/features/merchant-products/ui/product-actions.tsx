import {
  ArrowDownToLine,
  ArrowUpFromLine,
  Pencil,
  Trash2,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/shared/ui/button"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/shared/ui/tooltip"
import type { Product } from "../types"

export function ProductActions({
  product,
  onEdit,
  onToggleShelfStatus,
  onDelete,
}: {
  product: Product
  onEdit: (product: Product) => void
  onToggleShelfStatus: (product: Product) => void
  onDelete: (product: Product) => void
}) {
  const { t } = useTranslation("console")
  const shelfActionLabel =
    product.status === "on_shelf"
      ? t("products.actions.takeOffShelf")
      : t("products.actions.putOnShelf")

  return (
    <div className="flex items-center justify-end gap-2">
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={t("products.actions.edit")}
              onClick={() => onEdit(product)}
            />
          }
        >
          <Pencil aria-hidden />
        </TooltipTrigger>
        <TooltipContent>{t("products.actions.edit")}</TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={shelfActionLabel}
              onClick={() => onToggleShelfStatus(product)}
            />
          }
        >
          {product.status === "on_shelf" ? (
            <ArrowDownToLine aria-hidden />
          ) : (
            <ArrowUpFromLine aria-hidden />
          )}
        </TooltipTrigger>
        <TooltipContent>{shelfActionLabel}</TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={t("products.actions.delete")}
              className="text-destructive hover:bg-destructive/10 hover:text-destructive"
              onClick={() => onDelete(product)}
            />
          }
        >
          <Trash2 aria-hidden />
        </TooltipTrigger>
        <TooltipContent>{t("products.actions.delete")}</TooltipContent>
      </Tooltip>
    </div>
  )
}
