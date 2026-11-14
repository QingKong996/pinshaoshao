import { useMemo } from "react"
import {
  createColumnHelper,
  rowSortingFeature,
  tableFeatures,
  useTable,
  type OnChangeFn,
  type SortingState,
} from "@tanstack/react-table"
import { cn } from "cn"
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
import { defaultOrderForSort } from "../model/filters"
import type { Product } from "../types"
import { ProductActions } from "./product-actions"
import { ProductStatusBadge } from "./product-status-badge"

const productFeatures = tableFeatures({
  rowSortingFeature,
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})
const productColumnHelper = createColumnHelper<
  typeof productFeatures,
  Product
>()

export function ProductsTable({
  products,
  isPending,
  isCatalogEmpty,
  sorting,
  onSortingChange,
  productHref,
  onEdit,
  onToggleShelfStatus,
  onDelete,
}: {
  products: Product[]
  isPending: boolean
  isCatalogEmpty: boolean
  sorting: SortingState
  onSortingChange: OnChangeFn<SortingState>
  productHref: (productID: number) => string
  onEdit: (product: Product) => void
  onToggleShelfStatus: (product: Product) => void
  onDelete: (product: Product) => void
}) {
  const { t } = useTranslation("console")

  const columns = useMemo(
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
        productColumnHelper.accessor("total_stock", {
          header: t("products.table.stock"),
          sortDescFirst: defaultOrderForSort("total_stock") === "desc",
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
    [t, productHref, onEdit, onToggleShelfStatus, onDelete],
  )

  const table = useTable({
    features: productFeatures,
    columns,
    data: products,
    state: { sorting },
    onSortingChange,
    manualSorting: true,
    enableMultiSort: false,
    enableSortingRemoval: false,
    getRowId: (row) => String(row.id),
  })
  const rows = table.getRowModel().rows
  const columnCount = table.getAllLeafColumns().length

  return (
    <Table>
      <TableHeader>
        {table.getHeaderGroups().map((headerGroup) => (
          <TableRow key={headerGroup.id}>
            {headerGroup.headers.map((header) => {
              const meta = header.column.columnDef.meta
              const canSort = header.column.getCanSort()
              const sortDirection = header.column.getIsSorted()
              const headerLabel =
                typeof header.column.columnDef.header === "string"
                  ? header.column.columnDef.header
                  : undefined
              const ariaSort =
                sortDirection === "asc"
                  ? "ascending"
                  : sortDirection === "desc"
                    ? "descending"
                    : "none"

              return (
                <TableHead
                  key={header.id}
                  aria-sort={canSort ? ariaSort : undefined}
                  className={cn("px-4", meta?.headerClassName)}
                >
                  {header.isPlaceholder ? null : canSort ? (
                    <button
                      type="button"
                      className="inline-flex items-center gap-1 rounded-sm outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50"
                      aria-label={
                        headerLabel
                          ? t("products.table.sortBy", {
                              column: headerLabel,
                            })
                          : undefined
                      }
                      onClick={header.column.getToggleSortingHandler()}
                    >
                      <table.FlexRender header={header} />
                      {sortDirection === "asc" ? (
                        <ArrowUp className="size-3.5" aria-hidden />
                      ) : sortDirection === "desc" ? (
                        <ArrowDown className="size-3.5" aria-hidden />
                      ) : (
                        <ChevronsUpDown
                          className="size-3.5 text-muted-foreground/60"
                          aria-hidden
                        />
                      )}
                    </button>
                  ) : (
                    <table.FlexRender header={header} />
                  )}
                </TableHead>
              )
            })}
          </TableRow>
        ))}
      </TableHeader>
      <TableBody>
        {isPending ? (
          <ProductTableSkeleton />
        ) : rows.length === 0 ? (
          <TableRow>
            <TableCell
              colSpan={columnCount}
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
          rows.map((row) => (
            <TableRow key={row.id}>
              {row.getAllCells().map((cell) => (
                <TableCell
                  key={cell.id}
                  className={cn(
                    "px-4 py-3",
                    cell.column.columnDef.meta?.cellClassName,
                  )}
                >
                  <table.FlexRender cell={cell} />
                </TableCell>
              ))}
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
