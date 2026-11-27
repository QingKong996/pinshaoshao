import { useTable } from "@tanstack/react-table"
import { cn } from "cn"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Skeleton } from "@/shared/ui/skeleton"
import { TableEmptyState } from "@/shared/ui/table-empty-state"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import type { MerchantOrder } from "../types"
import {
  merchantOrderTableFeatures,
  useMerchantOrderColumns,
} from "./orders-table-columns"

export function OrdersTable({
  orders,
  isPending,
  isEmpty,
  renderRowActions,
}: {
  orders: MerchantOrder[]
  isPending: boolean
  isEmpty: boolean
  renderRowActions?: (order: MerchantOrder) => ReactNode
}) {
  const { t } = useTranslation("console")
  const columns = useMerchantOrderColumns(renderRowActions)
  const table = useTable({
    features: merchantOrderTableFeatures,
    columns,
    data: orders,
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
          <OrdersTableSkeleton />
        ) : rows.length === 0 ? (
          <TableEmptyState
            colSpan={columnCount}
            title={isEmpty ? t("orders.empty.title") : undefined}
            description={
              isEmpty
                ? t("orders.empty.description")
                : t("orders.empty.filtered")
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

function OrdersTableSkeleton() {
  return (
    <>
      {Array.from({ length: 6 }, (_, index) => (
        <TableRow key={index}>
          <TableCell className="px-4 py-3">
            <div className="flex flex-col gap-1.5">
              <Skeleton className="h-4 w-16" />
              <Skeleton className="h-3 w-24" />
            </div>
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-5 w-24" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-8" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-20" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-32" />
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}
