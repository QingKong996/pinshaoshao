import { useState } from "react"
import {
  CheckCircle2,
  CreditCard,
  LoaderCircle,
  ShieldCheck,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import { Button } from "@/shared/ui/button"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { QueryError } from "@/shared/ui/query-error"
import { formatDateTime, formatPriceCents } from "@/shared/lib/format"
import {
  CheckoutUnavailableError,
  PaymentClosedError,
  describeStartPaymentError,
} from "../model/errors"
import { useStartPayment } from "../model/mutations"

type OrderPaymentPanelProps = {
  token: string
  orderID: number
  orderStatus: "pending_payment" | "paid"
  amountCents: number
  expiresAt: string
  paidAt: string | null
  onPaymentSettled?: () => void
  onCheckout?: (checkoutURL: string) => void
}

function openCheckout(checkoutURL: string) {
  window.location.assign(checkoutURL)
}

export function OrderPaymentPanel({
  token,
  orderID,
  orderStatus,
  amountCents,
  expiresAt,
  paidAt,
  onPaymentSettled,
  onCheckout = openCheckout,
}: OrderPaymentPanelProps) {
  const { t } = useTranslation("payment")
  const startPayment = useStartPayment(token)
  const [settled, setSettled] = useState(false)
  const [localError, setLocalError] = useState<unknown>(null)
  const paid = orderStatus === "paid" || settled
  const error = localError ?? startPayment.error

  async function handlePay() {
    setLocalError(null)
    startPayment.reset()

    try {
      const result = await startPayment.mutateAsync(orderID)
      if (result.payment.status === "succeeded") {
        setSettled(true)
        onPaymentSettled?.()
        return
      }
      if (result.payment.status === "closed") {
        setLocalError(new PaymentClosedError())
        return
      }
      if (!result.checkoutURL) {
        setLocalError(new CheckoutUnavailableError())
        return
      }
      onCheckout(result.checkoutURL)
    } catch {
      // the mutation state renders the error
    }
  }

  function dismissError() {
    setLocalError(null)
    startPayment.reset()
  }

  return (
    <Card className="gap-4">
      <CardHeader className="flex-row items-center justify-between gap-3">
        <CardTitle>{t("panel.title")}</CardTitle>
        <Badge variant={paid ? "secondary" : "outline"}>
          {paid ? t("status.paid") : t("status.pending")}
        </Badge>
      </CardHeader>
      <CardContent className="space-y-4" aria-live="polite">
        {paid ? (
          <div className="space-y-3">
            <div className="flex items-start gap-3">
              <CheckCircle2
                className="mt-0.5 size-5 text-emerald-600 dark:text-emerald-400"
                aria-hidden
              />
              <div className="space-y-1">
                <p className="font-medium">{t("panel.paidTitle")}</p>
                <p className="text-sm text-muted-foreground">
                  {t("panel.paidDescription")}
                </p>
              </div>
            </div>
            {paidAt ? (
              <p className="text-sm text-muted-foreground">
                {t("panel.paidAt", { date: formatDateTime(paidAt) })}
              </p>
            ) : null}
          </div>
        ) : (
          <>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">
                {t("panel.amountDue")}
              </p>
              <p className="font-heading text-2xl font-semibold tracking-tight tabular-nums">
                {formatPriceCents(amountCents)}
              </p>
              <p className="text-sm text-muted-foreground">
                {t("panel.payBefore", {
                  date: formatDateTime(expiresAt),
                })}
              </p>
            </div>

            <Button
              type="button"
              size="lg"
              className="w-full"
              disabled={startPayment.isPending}
              aria-busy={startPayment.isPending}
              onClick={() => void handlePay()}
            >
              {startPayment.isPending ? (
                <LoaderCircle className="animate-spin" aria-hidden />
              ) : (
                <CreditCard aria-hidden />
              )}
              {startPayment.isPending
                ? t("panel.preparing")
                : t("panel.payNow")}
            </Button>

            <p className="flex items-start gap-2 text-xs leading-relaxed text-muted-foreground">
              <ShieldCheck className="mt-0.5 size-3.5 shrink-0" aria-hidden />
              {t("panel.redirectHint")}
            </p>
          </>
        )}
      </CardContent>

      {error ? (
        <CardContent>
          <QueryError
            title={t("errors.title")}
            description={describeStartPaymentError(error, t)}
            retryLabel={t("actions.dismiss")}
            onRetry={dismissError}
          />
        </CardContent>
      ) : null}
    </Card>
  )
}
