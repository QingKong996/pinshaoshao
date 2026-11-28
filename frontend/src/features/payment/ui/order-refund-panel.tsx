import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { formatDateTime, formatPriceCents } from "@/shared/lib/format"
import { useOrderPayment } from "../model/queries"
import type { RefundStatus } from "../types"

function refundBadgeVariant(status: RefundStatus) {
  switch (status) {
    case "succeeded":
      return "secondary"
    case "failed":
      return "destructive"
    default:
      return "outline"
  }
}

export function OrderRefundPanel({
  token,
  orderID,
}: {
  token: string
  orderID: number
}) {
  const { t } = useTranslation("payment")
  const paymentQuery = useOrderPayment(token, orderID)
  const refund = paymentQuery.data?.refund

  if (!refund) {
    return null
  }

  return (
    <Card className="gap-4">
      <CardHeader className="flex-row items-center justify-between gap-3">
        <CardTitle>{t("refund.title")}</CardTitle>
        <Badge variant={refundBadgeVariant(refund.status)}>
          {t(`refund.status.${refund.status}`)}
        </Badge>
      </CardHeader>
      <CardContent className="space-y-3 text-sm">
        <div className="flex items-center justify-between gap-4">
          <span className="text-muted-foreground">
            {t("refund.amount")}
          </span>
          <span className="font-medium tabular-nums">
            {formatPriceCents(refund.amount_cents)}
          </span>
        </div>
        {refund.refunded_at ? (
          <p className="text-muted-foreground">
            {t("refund.refundedAt", {
              date: formatDateTime(refund.refunded_at),
            })}
          </p>
        ) : null}
        {refund.failed_at ? (
          <p className="text-muted-foreground">
            {t("refund.failedAt", {
              date: formatDateTime(refund.failed_at),
            })}
          </p>
        ) : null}
      </CardContent>
    </Card>
  )
}
