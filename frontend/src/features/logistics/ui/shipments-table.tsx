import { useTable } from "@tanstack/react-table"
import { cn } from "cn"
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
import type { ShipmentSummary } from "../types"
import {
  shipmentTableFeatures,
  useShipmentColumns,
} from "./shipments-table-columns"

export function ShipmentsTable({
  shipments,
  isPending,
  isEmpty,
}: {
  shipments: ShipmentSummary[]
  isPending: boolean
  isEmpty: boolean
}) {
  const { t } = useTranslation("logistics")
  const columns = useShipmentColumns()
  const table = useTable({
    features: shipmentTableFeatures,
    columns,
    data: shipments,
    getRowId: (row) => row.shipment_no,
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
          <ShipmentsTableSkeleton />
        ) : rows.length === 0 ? (
          <TableEmptyState
            colSpan={columnCount}
            title={isEmpty ? t("merchant.empty.title") : undefined}
            description={
              isEmpty
                ? t("merchant.empty.description")
                : t("merchant.empty.filtered")
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

function ShipmentsTableSkeleton() {
  return (
    <>
      {Array.from({ length: 6 }, (_, index) => (
        <TableRow key={index}>
          <TableCell className="px-4 py-3">
            <div className="flex flex-col gap-1.5">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-3 w-20" />
            </div>
          </TableCell>
          <TableCell className="px-4 py-3">
            <div className="flex flex-col gap-1.5">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-3 w-28" />
            </div>
          </TableCell>
          <TableCell className="px-4 py-3">
            <div className="flex flex-col gap-1.5">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-3 w-32" />
            </div>
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-5 w-24" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-36" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-7 w-20" />
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}
