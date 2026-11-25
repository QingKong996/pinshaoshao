import { useMemo, type ReactNode } from "react"
import { createColumnHelper, tableFeatures } from "@tanstack/react-table"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import {
  formatDateTimeSeconds,
  formatPriceCents,
} from "@/shared/lib/format"
import type { MerchantOrder } from "../types"
import { MerchantOrderStatusBadge } from "./order-status-badge"

export const merchantOrderTableFeatures = tableFeatures({
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})

const merchantOrderColumnHelper = createColumnHelper<
  typeof merchantOrderTableFeatures,
  MerchantOrder
>()

export function useMerchantOrderColumns(
  renderRowActions?: (order: MerchantOrder) => ReactNode,
) {
  const { t } = useTranslation("console")

  return useMemo(
    () => {
      const columns = [
        merchantOrderColumnHelper.accessor("order_id", {
          header: t("orders.table.order"),
          cell: ({ row }) => (
            <div className="flex flex-col gap-0.5">
              <Link
                to={String(row.original.id)}
                className="w-fit font-medium text-primary tabular-nums hover:underline"
              >
                #{row.original.order_id}
              </Link>
              <span className="text-xs text-muted-foreground tabular-nums">
                {t("orders.table.subOrder", { id: row.original.id })}
              </span>
            </div>
          ),
        }),
        merchantOrderColumnHelper.display({
          id: "recipient",
          header: t("orders.table.recipient"),
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
        merchantOrderColumnHelper.accessor("shipping_address", {
          header: t("orders.table.shippingAddress"),
          meta: { cellClassName: "whitespace-normal" },
          cell: ({ getValue }) => (
            <span
              className="block max-w-64 truncate"
              title={getValue()}
            >
              {getValue()}
            </span>
          ),
        }),
        merchantOrderColumnHelper.accessor("status", {
          header: t("orders.table.status"),
          cell: ({ getValue }) => (
            <MerchantOrderStatusBadge status={getValue()} />
          ),
        }),
        merchantOrderColumnHelper.accessor("item_count", {
          header: t("orders.table.items"),
          meta: { cellClassName: "tabular-nums" },
        }),
        merchantOrderColumnHelper.accessor("total_amount_cents", {
          header: t("orders.table.total"),
          meta: { cellClassName: "tabular-nums" },
          cell: ({ getValue }) => formatPriceCents(getValue()),
        }),
        merchantOrderColumnHelper.accessor("created_at", {
          header: t("orders.table.createdAt"),
          meta: {
            cellClassName: "text-muted-foreground tabular-nums",
          },
          cell: ({ getValue }) => formatDateTimeSeconds(getValue()),
        }),
      ]
      if (renderRowActions) {
        columns.push(
          merchantOrderColumnHelper.display({
            id: "actions",
            header: t("orders.table.actions"),
            cell: ({ row }) => renderRowActions(row.original),
          }),
        )
      }
      return merchantOrderColumnHelper.columns(columns)
    },
    [renderRowActions, t],
  )
}
