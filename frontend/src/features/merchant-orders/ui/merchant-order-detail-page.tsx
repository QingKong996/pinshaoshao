import { ArrowLeft } from "lucide-react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import {
  formatDateTimeSeconds,
  formatPriceCents,
} from "@/shared/lib/format"
import { buttonVariants } from "@/shared/ui/button"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { Skeleton } from "@/shared/ui/skeleton"
import { MERCHANT_ORDER_NOT_FOUND_ERROR } from "../model/errors"
import { useMerchantOrder } from "../model/queries"
import type { MerchantOrderDetail, MerchantOrderItem } from "../types"
import { MerchantOrderStatusBadge } from "./order-status-badge"
import { QueryError } from "./query-error"

export default function MerchantOrderDetailPage({
  token,
  subOrderID,
  backHref,
  renderActions,
}: {
  token: string
  subOrderID: number
  backHref: string
  renderActions?: (order: MerchantOrderDetail) => ReactNode
}) {
  const { t } = useTranslation("console")
  const validSubOrderID = Number.isInteger(subOrderID) && subOrderID > 0
  const orderQuery = useMerchantOrder(token, subOrderID)

  useDocumentTitle(
    orderQuery.data
      ? t("documentTitle.orderDetail", { id: orderQuery.data.id })
      : t("documentTitle.orders"),
  )

  if (!validSubOrderID) {
    return (
      <div className="flex flex-col gap-6">
        <BackLink backHref={backHref} />
        <QueryError error={MERCHANT_ORDER_NOT_FOUND_ERROR} />
      </div>
    )
  }

  if (orderQuery.error && !orderQuery.data) {
    return (
      <div className="flex flex-col gap-6">
        <BackLink backHref={backHref} />
        <QueryError error={orderQuery.error} onRetry={orderQuery.refetch} />
      </div>
    )
  }

  if (orderQuery.isPending) {
    return <MerchantOrderDetailSkeleton />
  }

  if (!orderQuery.data) {
    return null
  }

  const order = orderQuery.data

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={
          <span className="flex flex-wrap items-center gap-2">
            {t("orders.detail.title", { id: order.id })}
            <MerchantOrderStatusBadge status={order.status} />
          </span>
        }
        description={t("orders.detail.description", {
          orderID: order.order_id,
          date: formatDateTimeSeconds(order.created_at),
        })}
        actions={
          <>
            <Link
              to={backHref}
              className={buttonVariants({ variant: "outline", size: "sm" })}
            >
              <ArrowLeft aria-hidden />
              {t("orders.detail.back")}
            </Link>
            {renderActions?.(order)}
          </>
        }
      />

      <Card>
        <CardHeader>
          <CardTitle>{t("orders.detail.summary.title")}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <SummaryItem
            label={t("orders.detail.summary.total")}
            value={formatPriceCents(order.total_amount_cents)}
          />
          <SummaryItem
            label={t("orders.detail.summary.items")}
            value={String(order.item_count)}
          />
          <SummaryItem
            label={t("orders.detail.summary.paidAt")}
            value={
              order.paid_at
                ? formatDateTimeSeconds(order.paid_at)
                : t("orders.detail.notPaid")
            }
          />
          <SummaryItem
            label={t("orders.detail.summary.createdAt")}
            value={formatDateTimeSeconds(order.created_at)}
          />
        </CardContent>
      </Card>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Card>
          <CardHeader>
            <CardTitle>{t("orders.detail.items.title")}</CardTitle>
          </CardHeader>
          <CardContent>
            {order.items.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("orders.detail.items.empty")}
              </p>
            ) : (
              <ul className="divide-y">
                {order.items.map((item) => (
                  <OrderItemRow key={item.id} item={item} />
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <div className="grid gap-4">
          <Card>
            <CardHeader>
              <CardTitle>{t("orders.detail.receiver.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <DetailItem
                label={t("orders.detail.receiver.name")}
                value={order.receiver_name}
              />
              <DetailItem
                label={t("orders.detail.receiver.phone")}
                value={order.receiver_phone}
              />
              <DetailItem
                label={t("orders.detail.receiver.address")}
                value={order.shipping_address}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>{t("orders.detail.meta.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <DetailItem
                label={t("orders.detail.meta.order")}
                value={`#${order.order_id}`}
              />
              <DetailItem
                label={t("orders.detail.meta.subOrder")}
                value={`#${order.id}`}
              />
              <DetailItem
                label={t("orders.detail.meta.updatedAt")}
                value={formatDateTimeSeconds(order.updated_at)}
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

function BackLink({ backHref }: { backHref: string }) {
  const { t } = useTranslation("console")

  return (
    <Link
      to={backHref}
      className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
    >
      <ArrowLeft className="size-4" aria-hidden />
      {t("orders.detail.back")}
    </Link>
  )
}

function OrderItemRow({ item }: { item: MerchantOrderItem }) {
  const { t } = useTranslation("console")

  return (
    <li className="flex items-start justify-between gap-4 py-3 first:pt-0 last:pb-0">
      <div className="min-w-0">
        <p className="font-medium">{item.product_name}</p>
        <p className="text-sm text-muted-foreground">
          {formatSpecs(item.sku_specs, t("orders.detail.items.noSpecs"))}
        </p>
        <p className="text-sm text-muted-foreground">
          {t("orders.detail.items.quantity", {
            price: formatPriceCents(item.unit_price_cents),
            quantity: item.quantity,
          })}
        </p>
      </div>
      <p className="shrink-0 font-medium tabular-nums">
        {formatPriceCents(item.line_amount_cents)}
      </p>
    </li>
  )
}

function formatSpecs(specs: Record<string, string>, fallback: string) {
  const values = Object.entries(specs)
  if (values.length === 0) {
    return fallback
  }
  return values.map(([name, value]) => `${name}: ${value}`).join(" · ")
}

function SummaryItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="space-y-1">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="font-medium tabular-nums">{value}</p>
    </div>
  )
}

function DetailItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-muted-foreground">{label}</p>
      <p className="font-medium break-words">{value}</p>
    </div>
  )
}

function MerchantOrderDetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-16 w-80" />
      <Skeleton className="h-32 rounded-xl" />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Skeleton className="h-72 rounded-xl" />
        <div className="grid gap-4">
          <Skeleton className="h-48 rounded-xl" />
          <Skeleton className="h-40 rounded-xl" />
        </div>
      </div>
    </div>
  )
}
