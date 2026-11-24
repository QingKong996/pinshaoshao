import {
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
import { Skeleton } from "@/shared/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import { TableEmptyState } from "@/shared/ui/table-empty-state"
import type { Product } from "../types"
import {
  productTableFeatures,
  useProductColumns,
} from "./product-table-columns"

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

  const columns = useProductColumns({
    productHref,
    onEdit,
    onToggleShelfStatus,
    onDelete,
  })

  const table = useTable({
    features: productTableFeatures,
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
          <TableEmptyState
            colSpan={columnCount}
            title={isCatalogEmpty ? t("products.empty.title") : undefined}
            description={
              isCatalogEmpty
                ? t("products.empty.description")
                : t("products.empty.filtered")
            }
          />
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
