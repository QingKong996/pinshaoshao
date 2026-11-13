import {
  ArrowDownToLine,
  ArrowUpFromLine,
  Boxes,
  MoreHorizontal,
  Pencil,
  Trash2,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Button } from "@/shared/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu"
import type { Product } from "../types"

export function ProductActionsMenu({
  product,
  productHref,
  onEdit,
  onToggleShelfStatus,
  onDelete,
}: {
  product: Product
  productHref: string
  onEdit: (product: Product) => void
  onToggleShelfStatus: (product: Product) => void
  onDelete: (product: Product) => void
}) {
  const { t } = useTranslation("console")

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("products.table.actions")}
          />
        }
      >
        <MoreHorizontal aria-hidden />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => onEdit(product)}>
          <Pencil aria-hidden />
          {t("products.actions.edit")}
        </DropdownMenuItem>
        <DropdownMenuItem render={<Link to={productHref} />}>
          <Boxes aria-hidden />
          {t("products.actions.manageSkus")}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => onToggleShelfStatus(product)}>
          {product.status === "on_shelf" ? (
            <ArrowDownToLine aria-hidden />
          ) : (
            <ArrowUpFromLine aria-hidden />
          )}
          {product.status === "on_shelf"
            ? t("products.actions.takeOffShelf")
            : t("products.actions.putOnShelf")}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          variant="destructive"
          onClick={() => onDelete(product)}
        >
          <Trash2 aria-hidden />
          {t("products.actions.delete")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
