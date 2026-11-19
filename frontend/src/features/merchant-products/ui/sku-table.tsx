import { useMemo } from "react"
import {
  createColumnHelper,
  tableFeatures,
  useTable,
} from "@tanstack/react-table"
import { cn } from "cn"
import { Pencil, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import { Button } from "@/shared/ui/button"
import { Skeleton } from "@/shared/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import { formatPriceCents } from "@/shared/lib/format"
import type { ProductSKU } from "../types"
import { formatSpecs } from "./format-specs"
import { TableEmptyState } from "./table-empty-state"

const skuFeatures = tableFeatures({
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})
const skuColumnHelper = createColumnHelper<typeof skuFeatures, ProductSKU>()

export function SKUTable({
  skus,
  isPending,
  showEmpty,
  onEdit,
  onDelete,
}: {
  skus: ProductSKU[]
  isPending: boolean
  showEmpty: boolean
  onEdit: (sku: ProductSKU) => void
  onDelete: (sku: ProductSKU) => void
}) {
  const { t } = useTranslation("console")

  const columns = useMemo(
    () =>
      skuColumnHelper.columns([
        skuColumnHelper.accessor("id", {
          id: "specs",
          header: t("productDetail.skus.table.specs"),
          meta: { cellClassName: "whitespace-normal" },
          cell: ({ row }) =>
            formatSpecs(
              row.original.specs,
              t("productDetail.skus.defaultSpecs"),
            ),
        }),
        skuColumnHelper.accessor("price_cents", {
          header: t("productDetail.skus.table.price"),
          meta: { cellClassName: "tabular-nums" },
          cell: ({ getValue }) => formatPriceCents(getValue()),
        }),
        skuColumnHelper.accessor("is_enabled", {
          header: t("productDetail.skus.table.enabled"),
          cell: ({ getValue }) => (
            <Badge variant={getValue() ? "secondary" : "outline"}>
              {getValue()
                ? t("productDetail.skus.enabled")
                : t("productDetail.skus.disabled")}
            </Badge>
          ),
        }),
        skuColumnHelper.display({
          id: "actions",
          header: t("productDetail.skus.table.actions"),
          meta: {
            headerClassName: "text-right",
            cellClassName: "text-right",
          },
          cell: ({ row }) => (
            <div className="flex justify-end gap-1">
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={t("productDetail.skus.actions.edit")}
                onClick={() => onEdit(row.original)}
              >
                <Pencil aria-hidden />
              </Button>
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={t("productDetail.skus.actions.delete")}
                onClick={() => onDelete(row.original)}
              >
                <Trash2 aria-hidden />
              </Button>
            </div>
          ),
        }),
      ]),
    [onDelete, onEdit, t],
  )

  const table = useTable({
    features: skuFeatures,
    columns,
    data: skus,
    getRowId: (row) => String(row.id),
  })
  const rows = table.getRowModel().rows
  const columnCount = table.getAllLeafColumns().length

  return (
    <Table>
      <TableHeader>
        {table.getHeaderGroups().map((headerGroup) => (
          <TableRow key={headerGroup.id}>
            {headerGroup.headers.map((header) => (
              <TableHead
                key={header.id}
                className={cn(
                  header.column.columnDef.meta?.headerClassName,
                )}
              >
                {header.isPlaceholder ? null : (
                  <table.FlexRender header={header} />
                )}
              </TableHead>
            ))}
          </TableRow>
        ))}
      </TableHeader>
      <TableBody>
        {isPending ? <SKUTableSkeleton /> : null}

        {showEmpty && rows.length === 0 ? (
          <TableEmptyState
            colSpan={columnCount}
            title={t("productDetail.skus.empty.title")}
            description={t("productDetail.skus.empty.description")}
            compact
          />
        ) : null}

        {rows.map((row) => (
          <TableRow key={row.id}>
            {row.getAllCells().map((cell) => (
              <TableCell
                key={cell.id}
                className={cn(cell.column.columnDef.meta?.cellClassName)}
              >
                <table.FlexRender cell={cell} />
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function SKUTableSkeleton() {
  return (
    <>
      {Array.from({ length: 3 }, (_, index) => (
        <TableRow key={index}>
          <TableCell>
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-16" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-5 w-16" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-6 w-16" />
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}
