import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import type { NotificationStatus, ShipmentStatus } from "../dto/shipment"

const SHIPMENT_STATUS_CLASS: Record<ShipmentStatus, string> = {
  pending_outbound:
    "border-amber-600/20 bg-amber-500/10 text-amber-700 dark:text-amber-300",
  pending_pickup:
    "border-orange-600/20 bg-orange-500/10 text-orange-700 dark:text-orange-300",
  picked_up:
    "border-sky-600/20 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  in_transit:
    "border-sky-600/20 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  at_transit_hub:
    "border-indigo-600/20 bg-indigo-500/10 text-indigo-700 dark:text-indigo-300",
  out_for_delivery:
    "border-violet-600/20 bg-violet-500/10 text-violet-700 dark:text-violet-300",
  delivered:
    "border-emerald-600/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
}

const NOTIFICATION_STATUS_CLASS: Record<NotificationStatus, string> = {
  pending:
    "border-amber-600/20 bg-amber-500/10 text-amber-700 dark:text-amber-300",
  delivering:
    "border-sky-600/20 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  succeeded:
    "border-emerald-600/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
  exhausted:
    "border-rose-600/20 bg-rose-500/10 text-rose-700 dark:text-rose-300",
}

export function ShipmentStatusBadge({ status }: { status: ShipmentStatus }) {
  const { t } = useTranslation("mocklogistics")
  return (
    <Badge variant="outline" className={SHIPMENT_STATUS_CLASS[status]}>
      {t(`status.${status}`)}
    </Badge>
  )
}

export function NotificationStatusBadge({
  status,
}: {
  status: NotificationStatus
}) {
  const { t } = useTranslation("mocklogistics")
  return (
    <Badge variant="outline" className={NOTIFICATION_STATUS_CLASS[status]}>
      {t(`notificationStatus.${status}`)}
    </Badge>
  )
}
