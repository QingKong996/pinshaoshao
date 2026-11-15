import { useMemo } from "react"
import {
  createColumnHelper,
  tableFeatures,
} from "@tanstack/react-table"
import { Pencil } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import {
  formatDateTimeSeconds,
  formatPriceCents,
} from "@/shared/lib/format"
import { Badge } from "@/shared/ui/badge"
import { Button } from "@/shared/ui/button"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/shared/ui/tooltip"
import type { InventoryItem } from "../types"
import { formatSpecs } from "./format-specs"

export const inventoryTableFeatures = tableFeatures({
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})
const inventoryColumnHelper = createColumnHelper<
  typeof inventoryTableFeatures,
  InventoryItem
>()

export function useInventoryColumns({
  productHref,
  onAdjustStock,
}: {
  productHref: (productID: number) => string
  onAdjustStock: (item: InventoryItem) => void
}) {
  const { t } = useTranslation("console")

  return useMemo(
    () =>
      inventoryColumnHelper.columns([
        inventoryColumnHelper.accessor("product_name", {
          header: t("inventory.table.product"),
          cell: ({ row }) => (
            <Link
              to={productHref(row.original.product_id)}
              className="font-medium hover:underline"
            >
              {row.original.product_name}
            </Link>
          ),
        }),
        inventoryColumnHelper.display({
          id: "specs",
          header: t("inventory.table.specs"),
          meta: { cellClassName: "whitespace-normal" },
          cell: ({ row }) =>
            formatSpecs(
              row.original.specs,
              t("inventory.defaultSpecs"),
            ),
        }),
        inventoryColumnHelper.accessor("category_name", {
          id: "category",
          header: t("inventory.table.category"),
          meta: { cellClassName: "text-muted-foreground" },
          cell: ({ getValue }) =>
            getValue() ?? t("inventory.uncategorized"),
        }),
        inventoryColumnHelper.accessor("price_cents", {
          header: t("inventory.table.price"),
          meta: { cellClassName: "tabular-nums" },
          cell: ({ getValue }) => formatPriceCents(getValue()),
        }),
        inventoryColumnHelper.accessor("stock", {
          header: t("inventory.table.stock"),
          meta: { cellClassName: "tabular-nums" },
          cell: ({ getValue }) => {
            const stock = getValue()
            return (
              <span
                className={
                  stock === 0
                    ? "font-medium text-destructive"
                    : undefined
                }
              >
                {stock}
              </span>
            )
          },
        }),
        inventoryColumnHelper.accessor("is_enabled", {
          header: t("inventory.table.status"),
          cell: ({ getValue }) => (
            <Badge variant={getValue() ? "secondary" : "outline"}>
              {getValue()
                ? t("inventory.enabled")
                : t("inventory.disabled")}
            </Badge>
          ),
        }),
        inventoryColumnHelper.accessor("updated_at", {
          header: t("inventory.table.updatedAt"),
          meta: {
            cellClassName: "text-muted-foreground tabular-nums",
          },
          cell: ({ getValue }) => formatDateTimeSeconds(getValue()),
        }),
        inventoryColumnHelper.display({
          id: "actions",
          header: t("inventory.table.actions"),
          meta: {
            headerClassName: "text-right",
            cellClassName: "text-right",
          },
          cell: ({ row }) => (
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={t("inventory.actions.adjustStock")}
                    onClick={() => onAdjustStock(row.original)}
                  />
                }
              >
                <Pencil aria-hidden />
              </TooltipTrigger>
              <TooltipContent>
                {t("inventory.actions.adjustStock")}
              </TooltipContent>
            </Tooltip>
          ),
        }),
      ]),
    [onAdjustStock, productHref, t],
  )
}
