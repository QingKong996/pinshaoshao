import { useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { Card } from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { TablePagination } from "@/shared/ui/table-pagination"
import {
  ALL,
  buildMerchantOrderListQuery,
  type StatusFilter,
} from "../model/filters"
import { useMerchantOrders } from "../model/queries"
import type { MerchantOrder } from "../types"
import { OrdersFilters } from "./orders-filters"
import { OrdersTable } from "./orders-table"
import { QueryError } from "./query-error"

const DEFAULT_PAGE_SIZE = 10
const EMPTY_ORDERS: MerchantOrder[] = []

export default function MerchantOrdersPage({
  token,
  renderRowActions,
}: {
  token: string
  renderRowActions?: (order: MerchantOrder) => ReactNode
}) {
  const { t } = useTranslation("console")
  useDocumentTitle(t("documentTitle.orders"))

  const [statusFilter, setStatusFilter] = useState<StatusFilter>(ALL)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)

  const ordersQuery = useMerchantOrders(
    token,
    buildMerchantOrderListQuery({ page, pageSize, statusFilter }),
  )
  const orders = ordersQuery.data?.items ?? EMPTY_ORDERS
  const total = ordersQuery.data?.total ?? 0
  const totalPages = ordersQuery.data?.total_pages ?? 0
  const hasFilters = statusFilter !== ALL

  function changeStatusFilter(nextStatus: StatusFilter) {
    setStatusFilter(nextStatus)
    setPage(1)
  }

  function changePageSize(nextPageSize: number) {
    setPageSize(nextPageSize)
    setPage(1)
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("orders.title")}
        description={t("orders.description")}
      />

      {ordersQuery.error && !ordersQuery.data ? (
        <QueryError
          error={ordersQuery.error}
          onRetry={ordersQuery.refetch}
        />
      ) : (
        <Card className="gap-0 overflow-hidden p-0">
          <OrdersFilters
            statusFilter={statusFilter}
            onStatusFilterChange={changeStatusFilter}
          />

          {ordersQuery.error ? (
            <div className="px-4 pt-4">
              <QueryError
                error={ordersQuery.error}
                onRetry={ordersQuery.refetch}
              />
            </div>
          ) : null}

          <OrdersTable
            orders={orders}
            isPending={ordersQuery.isPending}
            isEmpty={!hasFilters && total === 0}
            renderRowActions={renderRowActions}
          />

          <TablePagination
            page={page}
            pageSize={pageSize}
            total={total}
            totalPages={totalPages}
            onPageChange={setPage}
            onPageSizeChange={changePageSize}
          />
        </Card>
      )}
    </div>
  )
}
