import { useState } from "react"
import { ChevronLeft, ChevronRight, Package } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Button, buttonVariants } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { formatDateTime, formatPriceCents } from "@/shared/lib/format"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { describeOrderError } from "../model/errors"
import { useOrders } from "../model/queries"
import { OrderStatusBadge } from "./order-status-badge"

const PAGE_SIZE = 10

export function OrderListPage({ token }: { token: string }) {
  const { t } = useTranslation("order")
  useDocumentTitle(t("list.documentTitle"))

  const [page, setPage] = useState(1)
  const ordersQuery = useOrders(token, page, PAGE_SIZE)
  const result = ordersQuery.data

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("list.title")}
        description={t("list.description")}
      />

      {ordersQuery.error && !result ? (
        <QueryError
          title={t("errors.loadTitle")}
          description={describeOrderError(ordersQuery.error, t)}
          retryLabel={t("errors.retry")}
          onRetry={ordersQuery.refetch}
        />
      ) : ordersQuery.isPending && !result ? (
        <OrderListSkeleton />
      ) : !result || result.items.length === 0 ? (
        <Card className="items-center gap-3 px-6 py-14 text-center">
          <Package className="size-10 text-muted-foreground" aria-hidden />
          <div className="space-y-1">
            <h2 className="font-heading text-lg font-medium">
              {t("list.empty.title")}
            </h2>
            <p className="text-muted-foreground">
              {t("list.empty.description")}
            </p>
          </div>
          <Link to="/" className={buttonVariants()}>
            {t("list.empty.action")}
          </Link>
        </Card>
      ) : (
        <div className="grid gap-3">
          {result.items.map((order) => (
            <Link
              key={order.id}
              to={`/orders/${order.id}`}
              className="rounded-xl outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
            >
              <Card className="gap-3 px-4 py-4 transition-colors hover:bg-muted/40">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <h2 className="font-heading text-base font-medium">
                        {t("list.orderNumber", { id: order.id })}
                      </h2>
                      <OrderStatusBadge status={order.status} />
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {t("list.placedAt", {
                        date: formatDateTime(order.created_at),
                      })}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="font-heading text-lg font-semibold tabular-nums">
                      {formatPriceCents(order.total_amount_cents)}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {t("list.itemAndShopCount", {
                        items: order.item_count,
                        shops: order.shop_count,
                      })}
                    </p>
                  </div>
                </div>
              </Card>
            </Link>
          ))}
          <OrderListPagination
            page={result.page}
            pageSize={result.page_size}
            total={result.total}
            totalPages={result.total_pages}
            onPageChange={setPage}
          />
        </div>
      )}
    </div>
  )
}

function OrderListPagination({
  page,
  pageSize,
  total,
  totalPages,
  onPageChange,
}: {
  page: number
  pageSize: number
  total: number
  totalPages: number
  onPageChange: (page: number) => void
}) {
  const { t } = useTranslation("order")
  const visibleTotalPages = Math.max(1, totalPages)
  const visiblePage = Math.min(page, visibleTotalPages)
  const firstItem = total === 0 ? 0 : (visiblePage - 1) * pageSize + 1
  const lastItem = Math.min(visiblePage * pageSize, total)

  return (
    <div className="flex flex-col gap-3 border-t pt-4 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
      <p aria-live="polite">
        {t("list.pagination.summary", {
          start: firstItem,
          end: lastItem,
          total,
        })}
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="icon-sm"
          aria-label={t("list.pagination.previous")}
          disabled={visiblePage <= 1}
          onClick={() => onPageChange(visiblePage - 1)}
        >
          <ChevronLeft aria-hidden />
        </Button>
        <span className="min-w-20 text-center tabular-nums">
          {t("list.pagination.pageOf", {
            page: visiblePage,
            totalPages: visibleTotalPages,
          })}
        </span>
        <Button
          variant="outline"
          size="icon-sm"
          aria-label={t("list.pagination.next")}
          disabled={totalPages === 0 || visiblePage >= totalPages}
          onClick={() => onPageChange(visiblePage + 1)}
        >
          <ChevronRight aria-hidden />
        </Button>
      </div>
    </div>
  )
}

function OrderListSkeleton() {
  return (
    <div className="grid gap-3">
      {Array.from({ length: 3 }, (_, index) => (
        <Skeleton key={index} className="h-24 rounded-xl" />
      ))}
    </div>
  )
}
