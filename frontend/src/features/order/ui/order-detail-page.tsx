import { useState, type ReactNode } from "react"
import { LoaderCircle } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router"
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
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { formatDateTime, formatPriceCents } from "@/shared/lib/format"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { describeOrderError } from "../model/errors"
import { useCancelOrder } from "../model/mutations"
import { useOrder } from "../model/queries"
import type { Order } from "../types"
import { formatOrderSpecs } from "./format-specs"
import {
  OrderStatusBadge,
  SubOrderStatusBadge,
} from "./order-status-badge"

export function OrderDetailPage({
  token,
  orderID,
  renderSidebar,
}: {
  token: string
  orderID: number
  renderSidebar?: (order: Order) => ReactNode
}) {
  const { t } = useTranslation("order")
  useDocumentTitle(t("detail.documentTitle", { id: orderID }))

  const orderQuery = useOrder(token, orderID)
  const cancelOrder = useCancelOrder(token, orderID)
  const [cancelOpen, setCancelOpen] = useState(false)
  const navigate = useNavigate()

  async function confirmCancel() {
    try {
      await cancelOrder.mutateAsync()
      setCancelOpen(false)
    } catch {
      // the mutation state renders the error
    }
  }

  if (!Number.isInteger(orderID) || orderID <= 0) {
    return (
      <QueryError
        title={t("errors.notFoundTitle")}
        description={t("errors.notFound")}
        retryLabel={t("errors.backToOrders")}
        onRetry={() => navigate("/orders")}
      />
    )
  }

  if (orderQuery.error && !orderQuery.data) {
    return (
      <QueryError
        title={t("errors.loadTitle")}
        description={describeOrderError(orderQuery.error, t)}
        retryLabel={t("errors.retry")}
        onRetry={orderQuery.refetch}
      />
    )
  }

  if (orderQuery.isPending) {
    return <OrderDetailSkeleton />
  }

  if (!orderQuery.data) {
    return null
  }

  const order = orderQuery.data
  const canCancel = order.status === "pending_payment"
  const sidebar = renderSidebar?.(order)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("detail.title", { id: order.id })}
        description={t("detail.placedAt", {
          date: formatDateTime(order.created_at),
        })}
        actions={
          <>
            <Link
              to="/orders"
              className={buttonVariants({ variant: "outline", size: "sm" })}
            >
              {t("detail.backToOrders")}
            </Link>
            {canCancel ? (
              <Button
                variant="destructive"
                size="sm"
                onClick={() => setCancelOpen(true)}
              >
                {t("detail.cancel.action")}
              </Button>
            ) : null}
          </>
        }
      />

      {cancelOrder.error ? (
        <QueryError
          title={t("errors.cancelTitle")}
          description={describeOrderError(cancelOrder.error, t)}
          retryLabel={t("errors.dismiss")}
          onRetry={() => cancelOrder.reset()}
        />
      ) : null}

      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <CardTitle>{t("detail.summary.title")}</CardTitle>
          <OrderStatusBadge status={order.status} />
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-3">
          <SummaryItem
            label={t("detail.summary.total")}
            value={formatPriceCents(order.total_amount_cents)}
          />
          <SummaryItem
            label={t("detail.summary.expiresAt")}
            value={formatDateTime(order.expires_at)}
          />
          <SummaryItem
            label={t("detail.summary.status")}
            value={t(`status.${order.status}`)}
          />
        </CardContent>
      </Card>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <div className="grid gap-4">
          {order.sub_orders.map((subOrder) => (
            <Card key={subOrder.id}>
              <CardHeader className="flex-row items-start justify-between gap-4">
                <div className="space-y-1">
                  <CardTitle>{subOrder.shop_name}</CardTitle>
                  <p className="text-sm text-muted-foreground">
                    {t("detail.subOrder", { id: subOrder.id })}
                  </p>
                </div>
                <SubOrderStatusBadge status={subOrder.status} />
              </CardHeader>
              <CardContent className="space-y-0">
                <ul className="divide-y">
                  {subOrder.items.map((item) => (
                    <li
                      key={item.id}
                      className="flex items-start justify-between gap-4 py-3 first:pt-0 last:pb-0"
                    >
                      <div className="min-w-0">
                        <p className="font-medium">{item.product_name}</p>
                        <p className="text-sm text-muted-foreground">
                          {formatOrderSpecs(
                            item.sku_specs,
                            t("detail.noSpecs"),
                          )}
                        </p>
                        <p className="text-sm text-muted-foreground">
                          {t("detail.itemQuantity", {
                            quantity: item.quantity,
                            price: formatPriceCents(item.unit_price_cents),
                          })}
                        </p>
                      </div>
                      <p className="shrink-0 font-medium tabular-nums">
                        {formatPriceCents(item.line_amount_cents)}
                      </p>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          ))}
        </div>

        <aside
          className={
            sidebar
              ? "order-first grid gap-4 lg:order-none lg:sticky lg:top-24"
              : "grid gap-4 lg:sticky lg:top-24"
          }
        >
          {sidebar}
          <Card>
            <CardHeader>
              <CardTitle>{t("detail.receiver.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div>
                <p className="text-muted-foreground">
                  {t("detail.receiver.name")}
                </p>
                <p className="font-medium">{order.receiver_name}</p>
              </div>
              <div>
                <p className="text-muted-foreground">
                  {t("detail.receiver.phone")}
                </p>
                <p className="font-medium">{order.receiver_phone}</p>
              </div>
              <div>
                <p className="text-muted-foreground">
                  {t("detail.receiver.address")}
                </p>
                <p className="font-medium break-words">
                  {order.shipping_address}
                </p>
              </div>
            </CardContent>
          </Card>
        </aside>
      </div>

      <CancelOrderDialog
        order={order}
        open={cancelOpen}
        isPending={cancelOrder.isPending}
        onOpenChange={setCancelOpen}
        onConfirm={() => void confirmCancel()}
      />
    </div>
  )
}

function SummaryItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="space-y-1">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="font-medium tabular-nums">{value}</p>
    </div>
  )
}

function CancelOrderDialog({
  order,
  open,
  isPending,
  onOpenChange,
  onConfirm,
}: {
  order: Order
  open: boolean
  isPending: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  const { t } = useTranslation("order")

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
          <AlertDialogTitle>
            {t("detail.cancel.title")}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {t("detail.cancel.description", { id: order.id })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>
            {t("detail.cancel.keep")}
          </AlertDialogCancel>
          <Button
            variant="destructive"
            disabled={isPending}
            onClick={onConfirm}
          >
            {isPending ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : null}
            {t("detail.cancel.confirm")}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

function OrderDetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-16 w-64" />
      <Skeleton className="h-32 rounded-xl" />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Skeleton className="h-64 rounded-xl" />
        <Skeleton className="h-52 rounded-xl" />
      </div>
    </div>
  )
}
