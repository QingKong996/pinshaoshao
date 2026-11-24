import { useEffect, useState } from "react"
import {
  ArrowRight,
  ExternalLink,
  LoaderCircle,
  RefreshCw,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog"
import { Button, buttonVariants } from "@/shared/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { formatDateTime } from "@/shared/lib/format"
import { cn } from "@/shared/lib/utils"
import type { PaymentStatus, MockPayment } from "../dto/payment"
import { describeClosePaymentError } from "../model/errors"
import { useCloseMockPayment } from "../model/mutations"
import { MOCK_PAYMENT_ROUTE_PATHS } from "../model/paths"
import { CopyButton } from "./copy-button"
import { formatMockPaymentAmount } from "./format"
import { NotificationList } from "./notification-list"
import { PaymentStatusBadge } from "./payment-status-badge"

export function PaymentDetail({
  payment,
  isRefreshing,
  onRefresh,
}: {
  payment: MockPayment
  isRefreshing: boolean
  onRefresh: () => void
}) {
  const { t, i18n } = useTranslation("mockpayment")
  const closePayment = useCloseMockPayment(payment.payment_no)
  const [closeOpen, setCloseOpen] = useState(false)
  const [now, setNow] = useState(0)
  useEffect(() => {
    const updateNow = () => setNow(Date.now())
    updateNow()
    const interval = window.setInterval(updateNow, 5_000)
    return () => window.clearInterval(interval)
  }, [])

  const isExpired =
    payment.status === "pending" &&
    now > 0 &&
    Date.parse(payment.expires_at) <= now
  const locale = i18n.resolvedLanguage ?? i18n.language

  async function confirmClose() {
    try {
      await closePayment.mutateAsync()
      setCloseOpen(false)
    } catch {
      // mutation state renders the error
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={payment.subject}
        description={t("detail.subtitle", {
          merchantOrderNo: payment.merchant_order_no,
        })}
        actions={
          <>
            <Link
              to={MOCK_PAYMENT_ROUTE_PATHS.list}
              className={buttonVariants({ variant: "outline", size: "sm" })}
            >
              {t("actions.back")}
            </Link>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={isRefreshing}
              onClick={onRefresh}
            >
              <RefreshCw
                className={isRefreshing ? "animate-spin" : undefined}
                aria-hidden
              />
              {t("actions.refresh")}
            </Button>
          </>
        }
      />

      {closePayment.error ? (
        <QueryError
          title={t("errors.closeTitle")}
          description={describeClosePaymentError(closePayment.error, t)}
          retryLabel={t("actions.dismiss")}
          onRetry={() => closePayment.reset()}
        />
      ) : null}

      <Card className="gap-5">
        <CardContent className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <PaymentStatusBadge status={payment.status} />
              <span className="font-mono text-xs text-muted-foreground">
                {payment.payment_no}
              </span>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">
                {t("detail.amountLabel")}
              </p>
              <p className="font-heading text-3xl font-semibold tracking-tight tabular-nums sm:text-4xl">
                {formatMockPaymentAmount(
                  payment.amount_cents,
                  payment.currency,
                  locale,
                )}
              </p>
            </div>
            <p className="text-sm text-muted-foreground">
              {t("detail.expiresAt", {
                date: formatDateTime(payment.expires_at),
              })}
            </p>
            {isExpired ? (
              <p className="text-sm font-medium text-amber-700 dark:text-amber-300">
                {t("detail.expiredNotice")}
              </p>
            ) : null}
          </div>

          <div className="flex flex-wrap gap-2 lg:justify-end">
            <a
              href={payment.checkout_url}
              target="_blank"
              rel="noreferrer"
              className={buttonVariants()}
            >
              {t("actions.openCheckout")}
              <ExternalLink aria-hidden />
            </a>
            <CopyButton
              value={payment.checkout_url}
              label={t("actions.copyCheckout")}
            />
            {payment.status === "pending" ? (
              <Button
                type="button"
                variant="destructive"
                onClick={() => setCloseOpen(true)}
              >
                {t("actions.close")}
              </Button>
            ) : null}
          </div>
        </CardContent>
      </Card>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <PaymentInformation payment={payment} />
        <PaymentStateFlow status={payment.status} />
      </div>

      <NotificationList
        paymentNo={payment.payment_no}
        paymentStatus={payment.status}
      />

      <ClosePaymentDialog
        payment={payment}
        open={closeOpen}
        isPending={closePayment.isPending}
        onOpenChange={setCloseOpen}
        onConfirm={() => void confirmClose()}
      />
    </div>
  )
}

function PaymentInformation({ payment }: { payment: MockPayment }) {
  const { t } = useTranslation("mockpayment")

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("detail.fields.title")}</CardTitle>
        <CardDescription>{t("detail.fields.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-x-6 gap-y-5 sm:grid-cols-2">
          <PaymentField
            label={t("detail.fields.paymentNo")}
            value={payment.payment_no}
            mono
          />
          <PaymentField
            label={t("detail.fields.merchantOrderNo")}
            value={payment.merchant_order_no}
            mono
          />
          <PaymentField
            label={t("detail.fields.checkoutURL")}
            value={payment.checkout_url}
            mono
            className="sm:col-span-2"
          />
          {payment.return_url ? (
            <PaymentField
              label={t("detail.fields.returnURL")}
              value={payment.return_url}
              mono
              className="sm:col-span-2"
            />
          ) : null}
          <PaymentField
            label={t("detail.fields.createdAt")}
            value={formatDateTime(payment.created_at)}
          />
          <PaymentField
            label={t("detail.fields.expiresAt")}
            value={formatDateTime(payment.expires_at)}
          />
          {payment.paid_at ? (
            <PaymentField
              label={t("detail.fields.paidAt")}
              value={formatDateTime(payment.paid_at)}
            />
          ) : null}
          {payment.closed_at ? (
            <PaymentField
              label={t("detail.fields.closedAt")}
              value={formatDateTime(payment.closed_at)}
            />
          ) : null}
        </dl>
      </CardContent>
    </Card>
  )
}

function PaymentField({
  label,
  value,
  mono = false,
  className,
}: {
  label: string
  value: string
  mono?: boolean
  className?: string
}) {
  return (
    <div className={cn("min-w-0 space-y-1", className)}>
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd
        className={cn(
          "break-all font-medium",
          mono && "font-mono text-xs leading-relaxed",
        )}
      >
        {value}
      </dd>
    </div>
  )
}

function PaymentStateFlow({ status }: { status: PaymentStatus }) {
  const { t } = useTranslation("mockpayment")

  return (
    <Card className="lg:sticky lg:top-6">
      <CardHeader>
        <CardTitle>{t("detail.state.title")}</CardTitle>
        <CardDescription>{t("detail.state.description")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap items-center gap-2">
          <StateChip status="pending" active={status === "pending"} />
          <ArrowRight className="size-4 text-muted-foreground" aria-hidden />
          <div className="flex flex-wrap items-center gap-2">
            <StateChip status="paid" active={status === "paid"} />
            <span className="text-xs text-muted-foreground">
              {t("detail.state.or")}
            </span>
            <StateChip status="closed" active={status === "closed"} />
          </div>
        </div>
        <p className="text-xs leading-relaxed text-muted-foreground">
          {t("detail.state.note")}
        </p>
      </CardContent>
    </Card>
  )
}

function StateChip({
  status,
  active,
}: {
  status: PaymentStatus
  active: boolean
}) {
  const { t } = useTranslation("mockpayment")
  const activeClass = {
    pending:
      "border-amber-600/30 bg-amber-500/10 text-amber-800 dark:text-amber-200",
    paid: "border-emerald-600/30 bg-emerald-500/10 text-emerald-800 dark:text-emerald-200",
    closed: "border-foreground/20 bg-foreground text-background",
  }[status]

  return (
    <span
      aria-current={active ? "step" : undefined}
      className={cn(
        "rounded-full border px-2.5 py-1 text-xs font-medium",
        active
          ? activeClass
          : "border-border bg-background text-muted-foreground",
      )}
    >
      {t(`status.${status}`)}
    </span>
  )
}

function ClosePaymentDialog({
  payment,
  open,
  isPending,
  onOpenChange,
  onConfirm,
}: {
  payment: MockPayment
  open: boolean
  isPending: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  const { t } = useTranslation("mockpayment")

  return (
    <AlertDialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!isPending) {
          onOpenChange(nextOpen)
        }
      }}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t("detail.closeDialog.title")}</AlertDialogTitle>
          <AlertDialogDescription>
            {t("detail.closeDialog.description", {
              subject: payment.subject,
            })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>
            {t("detail.closeDialog.cancel")}
          </AlertDialogCancel>
          <Button variant="destructive" disabled={isPending} onClick={onConfirm}>
            {isPending ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : null}
            {t("detail.closeDialog.confirm")}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
