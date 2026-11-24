import { ArrowLeft, Package, UserRound } from "lucide-react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { buttonVariants } from "@/shared/ui/button"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { describeShipmentError } from "../model/errors"
import {
  useBuyerShipment,
  useMerchantShipment,
} from "../model/queries"
import { ShipmentTracking } from "./shipment-tracking"

export function ShipmentDetailPage({
  token,
  shipmentNo,
  scope,
  backTo,
  backLabel,
}: {
  token: string
  shipmentNo: string
  scope: "buyer" | "merchant"
  backTo: string
  backLabel: string
}) {
  const { t } = useTranslation("logistics")
  useDocumentTitle(t("documentTitle.detail"))
  const navigate = useNavigate()
  const buyerQuery = useBuyerShipment(
    token,
    shipmentNo,
    scope === "buyer",
  )
  const merchantQuery = useMerchantShipment(
    token,
    shipmentNo,
    scope === "merchant",
  )
  const shipmentQuery =
    scope === "buyer" ? buyerQuery : merchantQuery

  if (!shipmentNo) {
    return (
      <QueryError
        title={t("errors.notFoundTitle")}
        description={t("errors.notFound")}
        retryLabel={backLabel}
        onRetry={() => navigate(backTo)}
      />
    )
  }

  if (shipmentQuery.error && !shipmentQuery.data) {
    return (
      <QueryError
        title={t("errors.loadTitle")}
        description={describeShipmentError(shipmentQuery.error, t)}
        retryLabel={t("actions.retry")}
        onRetry={shipmentQuery.refetch}
      />
    )
  }

  if (shipmentQuery.isPending) {
    return <ShipmentDetailSkeleton />
  }

  if (!shipmentQuery.data) {
    return null
  }

  const shipment = shipmentQuery.data

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("detail.title")}
        description={t("detail.description", {
          shipmentNo: shipment.shipment_no,
          orderID: shipment.order_id,
        })}
        actions={
          <Link
            to={backTo}
            className={buttonVariants({ variant: "outline", size: "sm" })}
          >
            <ArrowLeft aria-hidden />
            {backLabel}
          </Link>
        }
      />

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Card>
          <CardHeader>
            <CardTitle>{t("detail.progress.title")}</CardTitle>
          </CardHeader>
          <CardContent>
            <ShipmentTracking shipment={shipment} />
          </CardContent>
        </Card>

        <div className="grid gap-4">
          <Card>
            <CardHeader>
              <CardTitle>{t("detail.receiver.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <DetailItem
                icon={<UserRound aria-hidden />}
                label={t("detail.receiver.name")}
                value={shipment.receiver.name}
              />
              <DetailItem
                label={t("detail.receiver.phone")}
                value={shipment.receiver.phone}
              />
              <DetailItem
                label={t("detail.receiver.address")}
                value={shipment.receiver.address}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>{t("detail.items.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <DetailItem
                icon={<Package aria-hidden />}
                label={t("detail.items.summary")}
                value={shipment.item_summary}
              />
              <DetailItem
                label={t("detail.items.subOrder")}
                value={`#${shipment.sub_order_id}`}
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

function DetailItem({
  icon,
  label,
  value,
}: {
  icon?: ReactNode
  label: string
  value: string
}) {
  return (
    <div className="flex items-start gap-2">
      {icon ? (
        <span className="mt-0.5 text-muted-foreground [&_svg]:size-3.5">
          {icon}
        </span>
      ) : null}
      <div className="min-w-0">
        <p className="text-muted-foreground">{label}</p>
        <p className="font-medium break-words">{value}</p>
      </div>
    </div>
  )
}

function ShipmentDetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-16 w-80" />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Skeleton className="h-96 rounded-xl" />
        <Skeleton className="h-64 rounded-xl" />
      </div>
    </div>
  )
}
