import {
  ArrowDown,
  ArrowUp,
  ChevronsUpDown,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import {
  formatDateTimeSeconds,
  formatPriceRange,
} from "@/shared/lib/format"
import { Skeleton } from "@/shared/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import type { Product, ProductSort, SortOrder } from "../types"
import { ProductActions } from "./product-actions"
import { ProductStatusBadge } from "./product-status-badge"

const COLUMN_COUNT = 8

export function ProductsTable({
  products,
  isPending,
  isCatalogEmpty,
  sort,
  order,
  onSort,
  productHref,
  onEdit,
  onToggleShelfStatus,
  onDelete,
}: {
  products: Product[]
  isPending: boolean
  isCatalogEmpty: boolean
  sort: ProductSort
  order: SortOrder
  onSort: (sort: ProductSort) => void
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
          <SortableTableHead
            column="name"
            label={t("products.table.product")}
            sort={sort}
            order={order}
            onSort={onSort}
          />
          <SortableTableHead
            column="category"
            label={t("products.table.category")}
            sort={sort}
            order={order}
            onSort={onSort}
          />
          <SortableTableHead
            column="status"
            label={t("products.table.status")}
            sort={sort}
            order={order}
            onSort={onSort}
          />
          <SortableTableHead
            column="sku_count"
            label={t("products.table.skus")}
            sort={sort}
            order={order}
            onSort={onSort}
          />
          <SortableTableHead
            column="total_stock"
            label={t("products.table.stock")}
            sort={sort}
            order={order}
            onSort={onSort}
          />
          <SortableTableHead
            column="min_price"
            label={t("products.table.price")}
            sort={sort}
            order={order}
            onSort={onSort}
          />
          <SortableTableHead
            column="created_at"
            label={t("products.table.createdAt")}
            sort={sort}
            order={order}
            onSort={onSort}
          />
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
              <TableCell className="px-4 py-3 text-muted-foreground tabular-nums">
                {formatDateTimeSeconds(product.created_at)}
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

function SortableTableHead({
  column,
  label,
  sort,
  order,
  onSort,
}: {
  column: ProductSort
  label: string
  sort: ProductSort
  order: SortOrder
  onSort: (sort: ProductSort) => void
}) {
  const { t } = useTranslation("console")
  const isActive = sort === column
  const ariaSort = isActive
    ? order === "asc"
      ? "ascending"
      : "descending"
    : "none"

  return (
    <TableHead className="px-4" aria-sort={ariaSort}>
      <button
        type="button"
        className="inline-flex items-center gap-1 rounded-sm outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50"
        aria-label={t("products.table.sortBy", { column: label })}
        onClick={() => onSort(column)}
      >
        {label}
        {isActive ? (
          order === "asc" ? (
            <ArrowUp className="size-3.5" aria-hidden />
          ) : (
            <ArrowDown className="size-3.5" aria-hidden />
          )
        ) : (
          <ChevronsUpDown
            className="size-3.5 text-muted-foreground/60"
            aria-hidden
          />
        )}
      </button>
    </TableHead>
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
            <Skeleton className="h-4 w-32" />
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
