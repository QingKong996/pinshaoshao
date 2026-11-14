import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { formatPriceRange } from "@/shared/lib/format"
import { Skeleton } from "@/shared/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import type { Product } from "../types"
import { ProductActions } from "./product-actions"
import { ProductStatusBadge } from "./product-status-badge"

const COLUMN_COUNT = 7

export function ProductsTable({
  products,
  isPending,
  isCatalogEmpty,
  productHref,
  onEdit,
  onToggleShelfStatus,
  onDelete,
}: {
  products: Product[]
  isPending: boolean
  isCatalogEmpty: boolean
  productHref: (productID: number) => string
  onEdit: (product: Product) => void
  onToggleShelfStatus: (product: Product) => void
  onDelete: (product: Product) => void
}) {
  const { t } = useTranslation("console")

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="px-4">
            {t("products.table.product")}
          </TableHead>
          <TableHead className="px-4">
            {t("products.table.category")}
          </TableHead>
          <TableHead className="px-4">
            {t("products.table.status")}
          </TableHead>
          <TableHead className="px-4">
            {t("products.table.skus")}
          </TableHead>
          <TableHead className="px-4">
            {t("products.table.stock")}
          </TableHead>
          <TableHead className="px-4">
            {t("products.table.price")}
          </TableHead>
          <TableHead className="px-4 text-right">
            {t("products.table.actions")}
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {isPending ? (
          <ProductTableSkeleton />
        ) : products.length === 0 ? (
          <TableRow>
            <TableCell
              colSpan={COLUMN_COUNT}
              className="px-4 py-16 text-center whitespace-normal"
            >
              {isCatalogEmpty ? (
                <>
                  <p className="font-medium">{t("products.empty.title")}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {t("products.empty.description")}
                  </p>
                </>
              ) : (
                <p className="text-sm text-muted-foreground">
                  {t("products.empty.filtered")}
                </p>
              )}
            </TableCell>
          </TableRow>
        ) : (
          products.map((product) => (
            <TableRow key={product.id}>
              <TableCell className="px-4 py-3">
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
              </TableCell>
              <TableCell className="px-4 py-3 text-muted-foreground">
                {product.category_name ??
                  t("products.filters.uncategorized")}
              </TableCell>
              <TableCell className="px-4 py-3">
                <ProductStatusBadge status={product.status} />
              </TableCell>
              <TableCell className="px-4 py-3 tabular-nums">
                {product.sku_count}
              </TableCell>
              <TableCell className="px-4 py-3 tabular-nums">
                {product.total_stock}
              </TableCell>
              <TableCell className="px-4 py-3 tabular-nums">
                {formatPriceRange(
                  product.min_price_cents,
                  product.max_price_cents,
                ) ?? "—"}
              </TableCell>
              <TableCell className="px-4 py-3 text-right">
                <ProductActions
                  product={product}
                  onEdit={onEdit}
                  onToggleShelfStatus={onToggleShelfStatus}
                  onDelete={onDelete}
                />
              </TableCell>
            </TableRow>
          ))
        )}
      </TableBody>
    </Table>
  )
}

function ProductTableSkeleton() {
  return (
    <>
      {Array.from({ length: 5 }, (_, index) => (
        <TableRow key={index}>
          <TableCell className="px-4 py-3">
            <div className="flex flex-col gap-1.5">
              <Skeleton className="h-4 w-40" />
              <Skeleton className="h-3 w-56" />
            </div>
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-20" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-5 w-16" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-8" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-8" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-24" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <div className="flex justify-end gap-2">
              <Skeleton className="size-7" />
              <Skeleton className="size-7" />
              <Skeleton className="size-7" />
            </div>
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}
