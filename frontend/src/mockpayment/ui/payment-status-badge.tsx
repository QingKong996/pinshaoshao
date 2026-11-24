import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import type {
  NotificationStatus,
  PaymentStatus,
} from "../dto/payment"

const PAYMENT_STATUS_CLASS: Record<PaymentStatus, string> = {
  pending:
    "border-amber-600/20 bg-amber-500/10 text-amber-700 dark:text-amber-300",
  paid: "border-emerald-600/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
  closed: "border-border bg-muted text-muted-foreground",
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

export function PaymentStatusBadge({ status }: { status: PaymentStatus }) {
  const { t } = useTranslation("mockpayment")
  return (
    <Badge variant="outline" className={PAYMENT_STATUS_CLASS[status]}>
      {t(`status.${status}`)}
    </Badge>
  )
}

export function NotificationStatusBadge({
  status,
}: {
  status: NotificationStatus
}) {
  const { t } = useTranslation("mockpayment")
  return (
    <Badge variant="outline" className={NOTIFICATION_STATUS_CLASS[status]}>
      {t(`notificationStatus.${status}`)}
    </Badge>
  )
}
