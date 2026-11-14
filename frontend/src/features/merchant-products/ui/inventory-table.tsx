import { useMemo } from "react"
import {
  createColumnHelper,
  tableFeatures,
  useTable,
} from "@tanstack/react-table"
import { cn } from "cn"
import { Pencil } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import {
  formatDateTimeSeconds,
  formatPriceCents,
} from "@/shared/lib/format"
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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/shared/ui/tooltip"
import type { InventoryItem } from "../types"

const inventoryFeatures = tableFeatures({
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})
const inventoryColumnHelper = createColumnHelper<
  typeof inventoryFeatures,
  InventoryItem
>()

export function InventoryTable({
  items,
  isPending,
  isEmpty,
  productHref,
  onAdjustStock,
}: {
  items: InventoryItem[]
  isPending: boolean
  isEmpty: boolean
  productHref: (productID: number) => string
  onAdjustStock: (item: InventoryItem) => void
}) {
  const { t } = useTranslation("console")

  const columns = useMemo(
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
            formatSpecs(row.original, t("inventory.defaultSpecs")),
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

  const table = useTable({
    features: inventoryFeatures,
    columns,
    data: items,
    getRowId: (row) => String(row.sku_id),
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
                  "px-4",
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
        {isPending ? (
          <InventoryTableSkeleton />
        ) : rows.length === 0 ? (
          <TableRow>
            <TableCell
              colSpan={columnCount}
              className="px-4 py-16 text-center whitespace-normal"
            >
              {isEmpty ? (
                <>
                  <p className="font-medium">{t("inventory.empty.title")}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {t("inventory.empty.description")}
                  </p>
                </>
              ) : (
                <p className="text-sm text-muted-foreground">
                  {t("inventory.empty.filtered")}
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

function formatSpecs(
  item: InventoryItem,
  emptyLabel: string,
) {
  const entries = Object.entries(item.specs).sort(([left], [right]) =>
    left.localeCompare(right),
  )
  if (entries.length === 0) {
    return emptyLabel
  }
  return entries.map(([key, value]) => `${key}: ${value}`).join(" / ")
}

function InventoryTableSkeleton() {
  return (
    <>
      {Array.from({ length: 6 }, (_, index) => (
        <TableRow key={index}>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-40" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-20" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-20" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-10" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-5 w-16" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <div className="flex justify-end">
              <Skeleton className="size-7" />
            </div>
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}
