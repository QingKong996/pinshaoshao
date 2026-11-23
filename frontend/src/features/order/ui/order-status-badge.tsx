import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import type { OrderStatus, SubOrderStatus } from "../types"

type BadgeVariant = "default" | "secondary" | "destructive" | "outline"

function orderVariant(status: OrderStatus): BadgeVariant {
  switch (status) {
    case "pending_payment":
      return "outline"
    case "completed":
      return "secondary"
    case "cancelled":
      return "destructive"
    default:
      return "default"
  }
}

function subOrderVariant(status: SubOrderStatus): BadgeVariant {
  switch (status) {
    case "pending_payment":
    case "awaiting_shipment":
      return "outline"
    case "completed":
      return "secondary"
    case "cancelled":
      return "destructive"
    default:
      return "default"
  }
}

export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  const { t } = useTranslation("order")
  return (
    <Badge variant={orderVariant(status)}>
      {t(`status.${status}`)}
    </Badge>
  )
}

export function SubOrderStatusBadge({ status }: { status: SubOrderStatus }) {
  const { t } = useTranslation("order")
  return (
    <Badge variant={subOrderVariant(status)}>
      {t(`subStatus.${status}`)}
    </Badge>
  )
}
