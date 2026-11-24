import { ExternalLink, Package, Truck } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { buttonVariants } from "@/shared/ui/button"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { describeShipmentError } from "../model/errors"
import { useOrderShipments } from "../model/queries"
import type { ShipmentSummary } from "../types"
import { ShipmentStatusBadge } from "./shipment-status-badge"

type SubOrderStatus = "awaiting_shipment" | "shipped" | "completed"

export function OrderShipmentPanel({
  token,
  orderID,
  subOrderID,
  subOrderStatus,
}: {
  token: string
  orderID: number
  subOrderID: number
  subOrderStatus: SubOrderStatus
}) {
  const { t } = useTranslation("logistics")
  const shipmentsQuery = useOrderShipments(
    token,
    orderID,
    subOrderStatus !== "awaiting_shipment",
  )
  const shipment = shipmentsQuery.data?.items.find(
    (item) => item.sub_order_id === subOrderID,
  )

  if (subOrderStatus === "awaiting_shipment") {
    return (
      <div className="mt-4 flex items-start gap-2 border-t pt-4 text-sm text-muted-foreground">
        <Package className="mt-0.5 size-4 shrink-0" aria-hidden />
        <p>{t("panel.awaitingShipment")}</p>
      </div>
    )
  }

  if (shipmentsQuery.isPending) {
    return (
      <div className="mt-4 space-y-2 border-t pt-4">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="h-3 w-64 max-w-full" />
      </div>
    )
  }

  if (shipmentsQuery.error) {
    return (
      <div className="mt-4 border-t pt-4">
        <QueryError
          title={t("errors.loadTitle")}
          description={describeShipmentError(shipmentsQuery.error, t)}
          retryLabel={t("actions.retry")}
          onRetry={shipmentsQuery.refetch}
        />
      </div>
    )
  }

  if (!shipment) {
    return (
      <div className="mt-4 flex items-start gap-2 border-t pt-4 text-sm text-muted-foreground">
        <Truck className="mt-0.5 size-4 shrink-0" aria-hidden />
        <p>{t("panel.trackingPending")}</p>
      </div>
    )
  }

  return (
    <div className="mt-4 space-y-3 border-t pt-4">
      <ShipmentSummaryRow shipment={shipment} />
      <Link
        to={`/shipments/${encodeURIComponent(shipment.shipment_no)}`}
        className={buttonVariants({ variant: "outline", size: "sm" })}
      >
        {t("actions.viewTracking")}
        <ExternalLink aria-hidden />
      </Link>
    </div>
  )
}

function ShipmentSummaryRow({ shipment }: { shipment: ShipmentSummary }) {
  const { t } = useTranslation("logistics")

  return (
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div className="min-w-0 space-y-1">
        <p className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
          <Truck className="size-4 text-muted-foreground" aria-hidden />
          <span className="font-medium">{shipment.carrier}</span>
          <span className="text-muted-foreground tabular-nums">
            {shipment.tracking_no}
          </span>
        </p>
        <p className="text-sm text-muted-foreground">
          {shipment.current_location || t("detail.summary.unknownLocation")}
        </p>
      </div>
      <ShipmentStatusBadge status={shipment.status} />
    </div>
  )
}
