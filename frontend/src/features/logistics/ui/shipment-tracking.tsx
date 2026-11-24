import { MapPin, PackageCheck, Truck } from "lucide-react"
import { useTranslation } from "react-i18next"
import { formatDateTime } from "@/shared/lib/format"
import type { Shipment } from "../types"
import { ShipmentStatusBadge } from "./shipment-status-badge"
import { ShipmentTimeline } from "./shipment-timeline"

export function ShipmentTracking({
  shipment,
  timelineLimit,
}: {
  shipment: Shipment
  timelineLimit?: number
}) {
  const { t } = useTranslation("logistics")

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="space-y-1">
          <p className="flex items-center gap-2 font-medium">
            <Truck className="size-4 text-muted-foreground" aria-hidden />
            {shipment.carrier}
            <span className="font-normal text-muted-foreground tabular-nums">
              {shipment.tracking_no}
            </span>
          </p>
          <p className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <MapPin className="size-3.5" aria-hidden />
            {shipment.current_location ||
              t("detail.summary.unknownLocation")}
          </p>
        </div>
        <ShipmentStatusBadge status={shipment.status} />
      </div>

      <div className="grid gap-3 text-sm sm:grid-cols-2">
        <div>
          <p className="text-muted-foreground">
            {t("detail.summary.shippedAt")}
          </p>
          <p className="font-medium tabular-nums">
            {formatDateTime(shipment.shipped_at)}
          </p>
        </div>
        <div>
          <p className="text-muted-foreground">
            {shipment.delivered_at
              ? t("detail.summary.deliveredAt")
              : t("detail.summary.estimatedRoute")}
          </p>
          <p className="flex items-center gap-1.5 font-medium">
            {shipment.delivered_at ? (
              <span className="tabular-nums">
                {formatDateTime(shipment.delivered_at)}
              </span>
            ) : (
              <>
                <PackageCheck
                  className="size-3.5 text-muted-foreground"
                  aria-hidden
                />
                {shipment.origin_city || t("detail.summary.originUnknown")}
                <span aria-hidden>→</span>
                {shipment.destination_city ||
                  t("detail.summary.destinationUnknown")}
              </>
            )}
          </p>
        </div>
      </div>

      <div className="space-y-3">
        <h3 className="text-sm font-medium">
          {timelineLimit
            ? t("panel.latestUpdates")
            : t("timeline.heading")}
        </h3>
        <ShipmentTimeline
          events={shipment.events}
          limit={timelineLimit}
        />
      </div>
    </div>
  )
}
