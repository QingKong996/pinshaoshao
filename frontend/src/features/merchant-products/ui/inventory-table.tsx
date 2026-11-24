import { useTable } from "@tanstack/react-table"
import { cn } from "cn"
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
import type { InventoryItem } from "../types"
import {
  inventoryTableFeatures,
  useInventoryColumns,
} from "./inventory-table-columns"

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

  const columns = useInventoryColumns({ productHref, onAdjustStock })

  const table = useTable({
    features: inventoryTableFeatures,
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
          <TableEmptyState
            colSpan={columnCount}
            title={isEmpty ? t("inventory.empty.title") : undefined}
            description={
              isEmpty
                ? t("inventory.empty.description")
                : t("inventory.empty.filtered")
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
            <Skeleton className="h-4 w-10" />
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
