import { useMemo } from "react"
import {
  createColumnHelper,
  rowSortingFeature,
  tableFeatures,
} from "@tanstack/react-table"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import {
  formatDateTimeSeconds,
  formatPriceRange,
} from "@/shared/lib/format"
import { defaultOrderForSort } from "../model/filters"
import type { Product } from "../types"
import { ProductActions } from "./product-actions"
import { ProductStatusBadge } from "./product-status-badge"

export const productTableFeatures = tableFeatures({
  rowSortingFeature,
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})
const productColumnHelper = createColumnHelper<
  typeof productTableFeatures,
  Product
>()

export function useProductColumns({
  productHref,
  onEdit,
  onToggleShelfStatus,
  onDelete,
}: {
  productHref: (productID: number) => string
  onEdit: (product: Product) => void
  onToggleShelfStatus: (product: Product) => void
  onDelete: (product: Product) => void
}) {
  const { t } = useTranslation("console")

  return useMemo(
    () =>
      productColumnHelper.columns([
        productColumnHelper.accessor("name", {
          header: t("products.table.product"),
          sortDescFirst: defaultOrderForSort("name") === "desc",
          cell: ({ row }) => {
            const product = row.original
            return (
              <div className="flex max-w-72 flex-col gap-0.5">
                <Link
                  to={productHref(product.id)}
                  className="truncate font-medium hover:underline"
                >
                  {product.name}
                </Link>
                {product.description ? (
                  <span className="truncate text-xs text-muted-foreground">
                    {product.description}
                  </span>
                ) : null}
              </div>
            )
          },
        }),
        productColumnHelper.accessor("category_name", {
          id: "category",
          header: t("products.table.category"),
          sortDescFirst: defaultOrderForSort("category") === "desc",
          meta: { cellClassName: "text-muted-foreground" },
          cell: ({ getValue }) =>
            getValue() ?? t("products.filters.uncategorized"),
        }),
        productColumnHelper.accessor("status", {
          header: t("products.table.status"),
          sortDescFirst: defaultOrderForSort("status") === "desc",
          cell: ({ getValue }) => (
            <ProductStatusBadge status={getValue()} />
          ),
        }),
        productColumnHelper.accessor("sku_count", {
          header: t("products.table.skus"),
          sortDescFirst: defaultOrderForSort("sku_count") === "desc",
          meta: { cellClassName: "tabular-nums" },
        }),
        productColumnHelper.accessor("min_price_cents", {
          id: "min_price",
          header: t("products.table.price"),
          sortDescFirst: defaultOrderForSort("min_price") === "desc",
          meta: { cellClassName: "tabular-nums" },
          cell: ({ row }) =>
            formatPriceRange(
              row.original.min_price_cents,
              row.original.max_price_cents,
            ) ?? "—",
        }),
        productColumnHelper.accessor("created_at", {
          header: t("products.table.createdAt"),
          sortDescFirst: defaultOrderForSort("created_at") === "desc",
          meta: {
            cellClassName: "text-muted-foreground tabular-nums",
          },
          cell: ({ getValue }) => formatDateTimeSeconds(getValue()),
        }),
        productColumnHelper.display({
          id: "actions",
          header: t("products.table.actions"),
          enableSorting: false,
          meta: {
            headerClassName: "text-right",
            cellClassName: "text-right",
          },
          cell: ({ row }) => (
            <ProductActions
              product={row.original}
              onEdit={onEdit}
              onToggleShelfStatus={onToggleShelfStatus}
              onDelete={onDelete}
            />
          ),
        }),
      ]),
    [onDelete, onEdit, onToggleShelfStatus, productHref, t],
  )
}
