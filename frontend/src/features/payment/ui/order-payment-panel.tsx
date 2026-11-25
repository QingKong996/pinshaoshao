import { useState } from "react"
import {
  CheckCircle2,
  CreditCard,
  LoaderCircle,
  ShieldCheck,
  XCircle,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog"
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
  describeClosePaymentError,
  describeStartPaymentError,
} from "../model/errors"
import { useClosePayment, useStartPayment } from "../model/mutations"

type OrderPaymentPanelProps = {
  token: string
  orderID: number
  orderStatus: "pending_payment" | "paid"
  amountCents: number
  expiresAt: string
  paidAt: string | null
  onPaymentSettled?: () => void
  onOrderCancelled?: () => void
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
  onOrderCancelled,
  onCheckout = openCheckout,
}: OrderPaymentPanelProps) {
  const { t } = useTranslation("payment")
  const startPayment = useStartPayment(token)
  const closePayment = useClosePayment(token)
  const [settled, setSettled] = useState(false)
  const [closed, setClosed] = useState(false)
  const [closeOpen, setCloseOpen] = useState(false)
  const [localError, setLocalError] = useState<unknown>(null)
  const paid = orderStatus === "paid" || settled
  const busy = startPayment.isPending || closePayment.isPending
  const error = localError ?? startPayment.error ?? closePayment.error

  async function handlePay() {
    setLocalError(null)
    closePayment.reset()
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

  async function handleClose() {
    setLocalError(null)
    startPayment.reset()
    closePayment.reset()

    try {
      const payment = await closePayment.mutateAsync(orderID)
      setCloseOpen(false)
      if (payment.status === "succeeded") {
        setSettled(true)
        onPaymentSettled?.()
        return
      }
      setClosed(true)
      onOrderCancelled?.()
    } catch {
      // the mutation state renders the error
      setCloseOpen(false)
    }
  }

  function dismissError() {
    setLocalError(null)
    startPayment.reset()
    closePayment.reset()
  }

  return (
    <Card className="gap-4">
      <CardHeader className="flex-row items-center justify-between gap-3">
        <CardTitle>{t("panel.title")}</CardTitle>
        <Badge
          variant={paid ? "secondary" : closed ? "destructive" : "outline"}
        >
          {paid
            ? t("status.paid")
            : closed
              ? t("status.cancelled")
              : t("status.pending")}
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
        ) : closed ? (
          <div className="flex items-start gap-3">
            <XCircle
              className="mt-0.5 size-5 text-destructive"
              aria-hidden
            />
            <div className="space-y-1">
              <p className="font-medium">{t("panel.closedTitle")}</p>
              <p className="text-sm text-muted-foreground">
                {t("panel.closedDescription")}
              </p>
            </div>
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
              disabled={busy}
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

            <Button
              type="button"
              variant="outline"
              className="w-full"
              disabled={busy}
              onClick={() => setCloseOpen(true)}
            >
              <XCircle aria-hidden />
              {t("panel.closeAction")}
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
            title={
              closePayment.error
                ? t("errors.closeTitle")
                : t("errors.title")
            }
            description={
              closePayment.error
                ? describeClosePaymentError(error, t)
                : describeStartPaymentError(error, t)
            }
            retryLabel={t("actions.dismiss")}
            onRetry={dismissError}
          />
        </CardContent>
      ) : null}

      <AlertDialog
        open={closeOpen}
        onOpenChange={(open) => {
          if (!busy) {
            setCloseOpen(open)
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("panel.closeDialog.title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("panel.closeDialog.description", { id: orderID })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={busy}>
              {t("panel.closeDialog.keep")}
            </AlertDialogCancel>
            <Button
              variant="destructive"
              disabled={busy}
              onClick={() => void handleClose()}
            >
              {closePayment.isPending ? (
                <LoaderCircle className="animate-spin" aria-hidden />
              ) : (
                <XCircle aria-hidden />
              )}
              {closePayment.isPending
                ? t("panel.closeDialog.closing")
                : t("panel.closeDialog.confirm")}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  )
}
