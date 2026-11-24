import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import type { MerchantOrderStatus } from "../types"

type BadgeVariant = "default" | "secondary" | "destructive" | "outline"

function statusVariant(status: MerchantOrderStatus): BadgeVariant {
  switch (status) {
    case "pending_payment":
      return "outline"
    case "awaiting_shipment":
      return "default"
    case "shipped":
    case "completed":
      return "secondary"
    case "cancelled":
      return "destructive"
  }
}

export function MerchantOrderStatusBadge({
  status,
}: {
  status: MerchantOrderStatus
}) {
  const { t } = useTranslation("console")
  return (
    <Badge variant={statusVariant(status)}>
      {t(`orders.status.${status}`)}
    </Badge>
  )
}
