import { useMemo } from "react"
import { createColumnHelper, tableFeatures } from "@tanstack/react-table"
import { ArrowUpRight } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { buttonVariants } from "@/shared/ui/button"
import { formatDateTimeSeconds } from "@/shared/lib/format"
import type { ShipmentSummary } from "../types"
import { ShipmentStatusBadge } from "./shipment-status-badge"

export const shipmentTableFeatures = tableFeatures({
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})

const shipmentColumnHelper = createColumnHelper<
  typeof shipmentTableFeatures,
  ShipmentSummary
>()

export function useShipmentColumns() {
  const { t } = useTranslation("logistics")

  return useMemo(
    () =>
      shipmentColumnHelper.columns([
        shipmentColumnHelper.accessor("shipment_no", {
          header: t("merchant.table.shipment"),
          cell: ({ row }) => (
            <div className="flex flex-col gap-0.5">
              <Link
                to={`/merchant/logistics/${encodeURIComponent(row.original.shipment_no)}`}
                className="font-medium underline-offset-4 hover:underline"
              >
                #{row.original.shipment_no.slice(0, 8)}
              </Link>
              <span className="text-xs text-muted-foreground tabular-nums">
                {t("merchant.table.subOrder", {
                  id: row.original.sub_order_id,
                })}
              </span>
            </div>
          ),
        }),
        shipmentColumnHelper.display({
          id: "recipient",
          header: t("merchant.table.recipient"),
          cell: ({ row }) => (
            <div className="flex flex-col gap-0.5">
              <span className="font-medium">
                {row.original.receiver_name}
              </span>
              <span className="text-xs text-muted-foreground tabular-nums">
                {row.original.receiver_phone}
              </span>
            </div>
          ),
        }),
        shipmentColumnHelper.display({
          id: "carrier",
          header: t("merchant.table.carrier"),
          cell: ({ row }) => (
            <div className="flex flex-col gap-0.5">
              <span className="font-medium">{row.original.carrier}</span>
              <span className="text-xs text-muted-foreground tabular-nums">
                {row.original.tracking_no}
              </span>
            </div>
          ),
        }),
        shipmentColumnHelper.accessor("status", {
          header: t("merchant.table.status"),
          cell: ({ getValue }) => (
            <ShipmentStatusBadge status={getValue()} />
          ),
        }),
        shipmentColumnHelper.accessor("current_location", {
          header: t("merchant.table.currentLocation"),
          meta: { cellClassName: "whitespace-normal" },
          cell: ({ getValue }) => (
            <span className="block max-w-48 truncate" title={getValue()}>
              {getValue() || t("detail.summary.unknownLocation")}
            </span>
          ),
        }),
        shipmentColumnHelper.accessor("shipped_at", {
          header: t("merchant.table.shippedAt"),
          meta: {
            cellClassName: "text-muted-foreground tabular-nums",
          },
          cell: ({ getValue }) => formatDateTimeSeconds(getValue()),
        }),
        shipmentColumnHelper.display({
          id: "actions",
          header: t("merchant.table.actions"),
          cell: ({ row }) => (
            <Link
              to={`/merchant/logistics/${encodeURIComponent(row.original.shipment_no)}`}
              className={buttonVariants({
                variant: "outline",
                size: "sm",
              })}
            >
              {t("actions.view")}
              <ArrowUpRight aria-hidden />
            </Link>
          ),
        }),
      ]),
    [t],
  )
}
