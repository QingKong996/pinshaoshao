import { useTranslation } from "react-i18next"
import { Link, useParams } from "react-router"
import { buttonVariants } from "@/shared/ui/button"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { describeShipmentError } from "../model/errors"
import { MOCK_LOGISTICS_ROUTE_PATHS } from "../model/paths"
import { useMockShipment } from "../model/queries"
import { MockLogisticsShell } from "./mock-logistics-shell"
import { ShipmentDetail } from "./shipment-detail"

export function MockLogisticsDetailPage() {
  const { shipmentNo = "" } = useParams()
  const { t } = useTranslation("mocklogistics")
  const shipment = useMockShipment(shipmentNo)
  useDocumentTitle(
    t("documentTitle.detail", {
      shipmentNo: shipment.data?.shipment_no ?? shipmentNo,
    }),
  )

  if (!shipmentNo) {
    return (
      <MockLogisticsShell>
        <QueryError
          title={t("errors.notFound")}
          description={t("errors.invalidShipment")}
          action={
            <Link
              to={MOCK_LOGISTICS_ROUTE_PATHS.list}
              className={buttonVariants({ variant: "outline", size: "sm" })}
            >
              {t("actions.back")}
            </Link>
          }
        />
      </MockLogisticsShell>
    )
  }

  if (shipment.error && !shipment.data) {
    return (
      <MockLogisticsShell>
        <div className="space-y-4">
          <QueryError
            title={t("errors.loadTitle")}
            description={describeShipmentError(shipment.error, t)}
            retryLabel={t("actions.retry")}
            onRetry={() => void shipment.refetch()}
          />
          <Link
            to={MOCK_LOGISTICS_ROUTE_PATHS.list}
            className={buttonVariants({ variant: "outline", size: "sm" })}
          >
            {t("actions.back")}
          </Link>
        </div>
      </MockLogisticsShell>
    )
  }

  if (shipment.isPending) {
    return (
      <MockLogisticsShell>
        <DetailSkeleton />
      </MockLogisticsShell>
    )
  }

  if (!shipment.data) {
    return null
  }

  return (
    <MockLogisticsShell>
      <ShipmentDetail
        shipment={shipment.data}
        isRefreshing={shipment.isFetching}
        onRefresh={() => void shipment.refetch()}
      />
    </MockLogisticsShell>
  )
}

function DetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-16 w-72" />
      <Skeleton className="h-44 rounded-xl" />
      <div className="grid gap-6 lg:grid-cols-2">
        <Skeleton className="h-72 rounded-xl" />
        <Skeleton className="h-72 rounded-xl" />
      </div>
      <Skeleton className="h-96 rounded-xl" />
      <Skeleton className="h-56 rounded-xl" />
    </div>
  )
}
