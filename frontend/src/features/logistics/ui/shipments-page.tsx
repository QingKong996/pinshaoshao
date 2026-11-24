import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { Card } from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { TablePagination } from "@/shared/ui/table-pagination"
import { describeShipmentError } from "../model/errors"
import {
  ALL,
  buildShipmentListQuery,
  type ShipmentStatusFilter,
} from "../model/filters"
import { useMerchantShipments } from "../model/queries"
import type { ShipmentSummary } from "../types"
import { ShipmentDispatchButton } from "./shipment-dispatch-button"
import { ShipmentsFilters } from "./shipments-filters"
import { ShipmentsTable } from "./shipments-table"

const DEFAULT_PAGE_SIZE = 10
const EMPTY_SHIPMENTS: ShipmentSummary[] = []

export default function MerchantShipmentsPage({
  token,
}: {
  token: string
}) {
  const { t } = useTranslation("logistics")
  useDocumentTitle(t("documentTitle.list"))

  const [statusFilter, setStatusFilter] =
    useState<ShipmentStatusFilter>(ALL)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)

  const shipmentsQuery = useMerchantShipments(
    token,
    buildShipmentListQuery({ page, pageSize, statusFilter }),
  )
  const shipments = shipmentsQuery.data?.items ?? EMPTY_SHIPMENTS
  const total = shipmentsQuery.data?.total ?? 0
  const totalPages = shipmentsQuery.data?.total_pages ?? 0
  const hasFilters = statusFilter !== ALL

  function changeStatusFilter(nextStatus: ShipmentStatusFilter) {
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
        title={t("merchant.title")}
        description={t("merchant.description")}
        actions={<ShipmentDispatchButton token={token} />}
      />

      {shipmentsQuery.error && !shipmentsQuery.data ? (
        <QueryError
          title={t("errors.loadTitle")}
          description={describeShipmentError(shipmentsQuery.error, t)}
          retryLabel={t("actions.retry")}
          onRetry={shipmentsQuery.refetch}
        />
      ) : (
        <Card className="gap-0 overflow-hidden p-0">
          <ShipmentsFilters
            statusFilter={statusFilter}
            onStatusFilterChange={changeStatusFilter}
          />

          {shipmentsQuery.error ? (
            <div className="px-4 pt-4">
              <QueryError
                title={t("errors.loadTitle")}
                description={describeShipmentError(
                  shipmentsQuery.error,
                  t,
                )}
                retryLabel={t("actions.retry")}
                onRetry={shipmentsQuery.refetch}
              />
            </div>
          ) : null}

          <ShipmentsTable
            shipments={shipments}
            isPending={shipmentsQuery.isPending}
            isEmpty={!hasFilters && total === 0}
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
