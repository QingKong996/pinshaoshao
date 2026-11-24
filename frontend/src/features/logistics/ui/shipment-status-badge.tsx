import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import type { ShipmentStatus } from "../types"

type BadgeVariant = "default" | "secondary" | "destructive" | "outline"

function statusVariant(status: ShipmentStatus): BadgeVariant {
  switch (status) {
    case "pending_pickup":
      return "outline"
    case "delivered":
      return "secondary"
    default:
      return "default"
  }
}

export function ShipmentStatusBadge({
  status,
}: {
  status: ShipmentStatus
}) {
  const { t } = useTranslation("logistics")
  return (
    <Badge variant={statusVariant(status)}>
      {t(`status.${status}`)}
    </Badge>
  )
}
