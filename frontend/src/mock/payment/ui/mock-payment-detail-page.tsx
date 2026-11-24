import { useTranslation } from "react-i18next"
import { Link, useParams } from "react-router"
import { buttonVariants } from "@/shared/ui/button"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { describePaymentError } from "../model/errors"
import { MOCK_PAYMENT_ROUTE_PATHS } from "../model/paths"
import { useMockPayment } from "../model/queries"
import { MockPaymentShell } from "./mock-payment-shell"
import { PaymentDetail } from "./payment-detail"

export function MockPaymentDetailPage() {
  const { paymentNo = "" } = useParams()
  const { t } = useTranslation("mockpayment")
  const payment = useMockPayment(paymentNo)
  useDocumentTitle(
    t("documentTitle.detail", {
      paymentNo: payment.data?.payment_no ?? paymentNo,
    }),
  )

  if (!paymentNo) {
    return (
      <MockPaymentShell>
        <QueryError
          title={t("errors.notFound")}
          description={t("errors.invalidPayment")}
          action={
            <Link
              to={MOCK_PAYMENT_ROUTE_PATHS.list}
              className={buttonVariants({ variant: "outline", size: "sm" })}
            >
              {t("actions.back")}
            </Link>
          }
        />
      </MockPaymentShell>
    )
  }

  if (payment.error && !payment.data) {
    return (
      <MockPaymentShell>
        <div className="space-y-4">
          <QueryError
            title={t("errors.loadTitle")}
            description={describePaymentError(payment.error, t)}
            retryLabel={t("actions.retry")}
            onRetry={() => void payment.refetch()}
          />
          <Link
            to={MOCK_PAYMENT_ROUTE_PATHS.list}
            className={buttonVariants({ variant: "outline", size: "sm" })}
          >
            {t("actions.back")}
          </Link>
        </div>
      </MockPaymentShell>
    )
  }

  if (payment.isPending) {
    return (
      <MockPaymentShell>
        <DetailSkeleton />
      </MockPaymentShell>
    )
  }

  if (!payment.data) {
    return null
  }

  return (
    <MockPaymentShell>
      <PaymentDetail
        payment={payment.data}
        isRefreshing={payment.isFetching}
        onRefresh={() => void payment.refetch()}
      />
    </MockPaymentShell>
  )
}

function DetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-16 w-72" />
      <Skeleton className="h-40 rounded-xl" />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Skeleton className="h-80 rounded-xl" />
        <Skeleton className="h-52 rounded-xl" />
      </div>
      <Skeleton className="h-56 rounded-xl" />
    </div>
  )
}
