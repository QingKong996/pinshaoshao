import {
  CheckCircle2,
  LoaderCircle,
  MapPin,
  Package,
  RefreshCw,
  Store,
  Truck,
} from "lucide-react"
import type { ParseKeys } from "i18next"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
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
import type {
  MockShipment,
  ShipmentParty,
  ShipmentStatus,
} from "../dto/shipment"
import { describeShipmentActionError } from "../model/errors"
import {
  useAdvanceMockShipment,
  useCompleteMockShipment,
  useOutboundMockShipment,
} from "../model/mutations"
import { MOCK_LOGISTICS_ROUTE_PATHS } from "../model/paths"
import { NotificationList } from "./notification-list"
import { ShipmentStatusBadge } from "./shipment-status-badge"

type ShipmentAction = "outbound" | "advance" | "complete"

export function ShipmentDetail({
  shipment,
  isRefreshing,
  onRefresh,
}: {
  shipment: MockShipment
  isRefreshing: boolean
  onRefresh: () => void
}) {
  const { t } = useTranslation("mocklogistics")
  const outbound = useOutboundMockShipment(shipment.shipment_no)
  const advance = useAdvanceMockShipment(shipment.shipment_no)
  const complete = useCompleteMockShipment(shipment.shipment_no)
  const action = nextAction(shipment.status)
  const actionMutation = {
    outbound,
    advance,
    complete,
  }[action ?? "outbound"]
  const isActing =
    outbound.isPending || advance.isPending || complete.isPending
  const actionError = outbound.error ?? advance.error ?? complete.error

  function dismissActionError() {
    outbound.reset()
    advance.reset()
    complete.reset()
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={shipment.item_summary}
        description={t("detail.subtitle", {
          merchantOrderNo: shipment.merchant_order_no,
        })}
        actions={
          <>
            <Link
              to={MOCK_LOGISTICS_ROUTE_PATHS.list}
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

      {actionError ? (
        <QueryError
          title={t("errors.actionTitle")}
          description={describeShipmentActionError(actionError, t)}
          retryLabel={t("actions.dismiss")}
          onRetry={dismissActionError}
        />
      ) : null}

      <Card className="gap-5">
        <CardContent className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-2">
              <ShipmentStatusBadge status={shipment.status} />
              <span className="font-mono text-xs text-muted-foreground">
                {shipment.tracking_no}
              </span>
            </div>
            <div className="grid gap-4 sm:grid-cols-3">
              <SummaryField
                label={t("detail.carrier")}
                value={shipment.carrier}
              />
              <SummaryField
                label={t("detail.currentLocation")}
                value={shipment.current_location}
              />
              <SummaryField
                label={t("detail.route")}
                value={t("detail.routeValue", {
                  origin: shipment.origin_city,
                  destination: shipment.destination_city,
                })}
              />
            </div>
          </div>

          <div className="flex flex-wrap gap-2 lg:justify-end">
            {action ? (
              <Button
                type="button"
                disabled={isActing}
                onClick={() => {
                  void actionMutation.mutateAsync().catch(() => {
                    // mutation state renders the error
                  })
                }}
              >
                {isActing ? (
                  <LoaderCircle className="animate-spin" aria-hidden />
                ) : action === "outbound" ? (
                  <Store aria-hidden />
                ) : action === "advance" ? (
                  <Truck aria-hidden />
                ) : (
                  <CheckCircle2 aria-hidden />
                )}
                {isActing
                  ? t("actions.processing")
                  : t(`actions.${action}`)}
              </Button>
            ) : (
              <span className="flex items-center gap-2 rounded-lg bg-muted px-3 py-2 text-sm text-muted-foreground">
                <CheckCircle2 className="size-4" aria-hidden />
                {t("detail.actions.terminalHint")}
              </span>
            )}
          </div>
        </CardContent>
      </Card>

      <div className="grid items-start gap-6 lg:grid-cols-2">
        <ShipmentInformation shipment={shipment} />
        <ShipmentParties
          sender={shipment.sender}
          receiver={shipment.receiver}
        />
      </div>

      <ActionHint status={shipment.status} />
      <ShipmentTimeline shipment={shipment} />
      <NotificationList
        shipmentNo={shipment.shipment_no}
        isActive={
          shipment.status !== "pending_outbound" &&
          shipment.status !== "delivered"
        }
      />
    </div>
  )
}

function nextAction(status: ShipmentStatus): ShipmentAction | null {
  switch (status) {
    case "pending_outbound":
      return "outbound"
    case "pending_pickup":
    case "picked_up":
    case "in_transit":
    case "at_transit_hub":
      return "advance"
    case "out_for_delivery":
      return "complete"
    case "delivered":
      return null
  }
}

function SummaryField({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0 space-y-1">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="break-words text-sm font-medium">{value}</p>
    </div>
  )
}

function ShipmentInformation({ shipment }: { shipment: MockShipment }) {
  const { t } = useTranslation("mocklogistics")

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("detail.fields.title")}</CardTitle>
        <CardDescription>{t("detail.fields.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-x-6 gap-y-5 sm:grid-cols-2">
          <DetailField
            label={t("detail.fields.shipmentNo")}
            value={shipment.shipment_no}
            mono
          />
          <DetailField
            label={t("detail.fields.trackingNo")}
            value={shipment.tracking_no}
            mono
          />
          <DetailField
            label={t("detail.fields.merchantOrderNo")}
            value={shipment.merchant_order_no}
            mono
          />
          <DetailField
            label={t("detail.fields.carrier")}
            value={shipment.carrier}
          />
          <DetailField
            label={t("detail.fields.itemSummary")}
            value={shipment.item_summary}
            className="sm:col-span-2"
          />
          <DetailField
            label={t("detail.fields.notifyURL")}
            value={shipment.notify_url}
            mono
            className="sm:col-span-2"
          />
          <DetailField
            label={t("detail.fields.createdAt")}
            value={formatDateTime(shipment.created_at)}
          />
          <DetailField
            label={t("detail.fields.updatedAt")}
            value={formatDateTime(shipment.updated_at)}
          />
          {shipment.dispatched_at ? (
            <DetailField
              label={t("detail.fields.dispatchedAt")}
              value={formatDateTime(shipment.dispatched_at)}
            />
          ) : null}
          {shipment.picked_up_at ? (
            <DetailField
              label={t("detail.fields.pickedUpAt")}
              value={formatDateTime(shipment.picked_up_at)}
            />
          ) : null}
          {shipment.delivered_at ? (
            <DetailField
              label={t("detail.fields.deliveredAt")}
              value={formatDateTime(shipment.delivered_at)}
            />
          ) : null}
        </dl>
      </CardContent>
    </Card>
  )
}

function DetailField({
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

function ShipmentParties({
  sender,
  receiver,
}: {
  sender: ShipmentParty
  receiver: ShipmentParty
}) {
  const { t } = useTranslation("mocklogistics")

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("detail.parties.title")}</CardTitle>
        <CardDescription>{t("detail.parties.description")}</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        <PartyBlock title={t("detail.parties.sender")} party={sender} />
        <PartyBlock title={t("detail.parties.receiver")} party={receiver} />
      </CardContent>
    </Card>
  )
}

function PartyBlock({ title, party }: { title: string; party: ShipmentParty }) {
  const { t } = useTranslation("mocklogistics")

  return (
    <div className="rounded-xl border bg-muted/20 p-4">
      <div className="mb-3 flex items-center gap-2">
        <Package className="size-4 text-muted-foreground" aria-hidden />
        <p className="text-sm font-medium">{title}</p>
      </div>
      <dl className="grid gap-3 sm:grid-cols-2">
        <div className="space-y-1">
          <dt className="text-xs text-muted-foreground">
            {t("detail.parties.phone")}
          </dt>
          <dd className="text-sm font-medium">
            {party.name} · {party.phone}
          </dd>
        </div>
        <div className="min-w-0 space-y-1 sm:col-span-2">
          <dt className="text-xs text-muted-foreground">
            {t("detail.parties.address")}
          </dt>
          <dd className="break-words text-sm font-medium">{party.address}</dd>
        </div>
      </dl>
    </div>
  )
}

function ActionHint({ status }: { status: ShipmentStatus }) {
  const { t } = useTranslation("mocklogistics")
  const hints: Record<ShipmentStatus, ParseKeys<"mocklogistics">> = {
    pending_outbound: "detail.actions.outboundHint",
    pending_pickup: "detail.actions.advanceHint",
    picked_up: "detail.actions.advanceHint",
    in_transit: "detail.actions.advanceHint",
    at_transit_hub: "detail.actions.advanceHint",
    out_for_delivery: "detail.actions.completeHint",
    delivered: "detail.actions.terminalHint",
  }

  return (
    <div className="flex items-start gap-3 rounded-xl border border-dashed bg-background p-4">
      <Store className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />
      <div>
        <p className="text-sm font-medium">{t("detail.actions.title")}</p>
        <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
          {t(hints[status])}
        </p>
      </div>
    </div>
  )
}

function ShipmentTimeline({ shipment }: { shipment: MockShipment }) {
  const { t } = useTranslation("mocklogistics")
  const events = [...shipment.events].reverse()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("detail.timeline.title")}</CardTitle>
        <CardDescription>{t("detail.timeline.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        {events.length === 0 ? (
          <div className="rounded-xl border border-dashed px-4 py-10 text-center text-sm text-muted-foreground">
            {t("detail.timeline.empty")}
          </div>
        ) : (
          <ol className="relative ml-3 border-l pl-6">
            {events.map((event, index) => (
              <li
                key={event.sequence}
                className={index === 0 ? "pb-0" : "pb-6"}
              >
                <span
                  className={cn(
                    "absolute -left-1.5 mt-1.5 size-3 rounded-full border-2 border-background",
                    index === 0 ? "bg-primary" : "bg-muted-foreground/40",
                  )}
                  aria-hidden
                />
                <div className="flex flex-wrap items-center gap-2">
                  <ShipmentStatusBadge status={event.status} />
                  <span className="text-xs text-muted-foreground">
                    {t("detail.timeline.sequence", {
                      sequence: event.sequence,
                    })}
                  </span>
                  <time className="ml-auto text-xs text-muted-foreground">
                    {formatDateTime(event.occurred_at)}
                  </time>
                </div>
                <p className="mt-2 text-sm font-medium">{event.description}</p>
                <p className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                  <MapPin className="size-3.5" aria-hidden />
                  {event.location}
                </p>
              </li>
            ))}
          </ol>
        )}
      </CardContent>
    </Card>
  )
}
