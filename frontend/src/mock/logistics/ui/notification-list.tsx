import { LoaderCircle, RefreshCw, RotateCcw } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/shared/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { formatDateTime } from "@/shared/lib/format"
import type { MockShipmentNotification } from "../dto/shipment"
import { describeNotificationError } from "../model/errors"
import { useRetryMockShipmentNotification } from "../model/mutations"
import { useMockShipmentNotifications } from "../model/queries"
import { NotificationStatusBadge } from "./shipment-status-badge"

export function NotificationList({ shipmentNo }: { shipmentNo: string }) {
  const { t } = useTranslation("mocklogistics")
  const notifications = useMockShipmentNotifications(shipmentNo)
  const retryNotification = useRetryMockShipmentNotification(shipmentNo)
  const items = notifications.data?.items ?? []

  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between gap-4">
        <div className="space-y-1">
          <CardTitle>{t("detail.notifications.title")}</CardTitle>
          <CardDescription>
            {t("detail.notifications.description")}
          </CardDescription>
        </div>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          disabled={notifications.isFetching}
          onClick={() => void notifications.refetch()}
        >
          <RefreshCw
            className={notifications.isFetching ? "animate-spin" : undefined}
            aria-hidden
          />
          {t("actions.refresh")}
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        {retryNotification.error ? (
          <QueryError
            title={t("errors.retryTitle")}
            description={describeNotificationError(retryNotification.error, t)}
            retryLabel={t("actions.dismiss")}
            onRetry={() => retryNotification.reset()}
          />
        ) : null}

        {notifications.isPending ? (
          <div className="space-y-3">
            <Skeleton className="h-32 rounded-xl" />
            <Skeleton className="h-32 rounded-xl" />
          </div>
        ) : notifications.error ? (
          <QueryError
            title={t("errors.notificationsTitle")}
            description={t("errors.loadNotifications")}
            retryLabel={t("actions.retry")}
            onRetry={() => void notifications.refetch()}
          />
        ) : items.length === 0 ? (
          <div className="rounded-xl border border-dashed px-4 py-10 text-center text-sm text-muted-foreground">
            {t("detail.notifications.empty")}
          </div>
        ) : (
          <ul className="grid gap-4">
            {items.map((notification) => (
              <NotificationItem
                key={notification.id}
                notification={notification}
                isRetrying={retryNotification.isPending}
                onRetry={() => retryNotification.mutate(notification.id)}
              />
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}

function NotificationItem({
  notification,
  isRetrying,
  onRetry,
}: {
  notification: MockShipmentNotification
  isRetrying: boolean
  onRetry: () => void
}) {
  const { t } = useTranslation("mocklogistics")
  const showNextAttempt =
    notification.status === "pending" || notification.status === "delivering"

  return (
    <li className="rounded-xl border bg-muted/20 p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0 space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <NotificationStatusBadge status={notification.status} />
            <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
              {notification.event_type}
            </code>
          </div>
          <p className="break-all font-mono text-xs text-muted-foreground">
            {t("detail.notifications.eventId")}: {notification.event_id}
          </p>
        </div>
        {notification.status === "exhausted" ? (
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={isRetrying}
            onClick={onRetry}
          >
            {isRetrying ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : (
              <RotateCcw aria-hidden />
            )}
            {isRetrying ? t("actions.retrying") : t("actions.retry")}
          </Button>
        ) : null}
      </div>

      <dl className="mt-4 grid gap-4 border-t pt-4 sm:grid-cols-2">
        <NotificationField
          label={t("detail.notifications.targetURL")}
          value={notification.target_url}
          mono
        />
        <NotificationField
          label={t("detail.notifications.attempts")}
          value={t("detail.notifications.attemptValue", {
            attempts: notification.attempt_count,
            max: notification.max_attempts,
          })}
        />
        {showNextAttempt ? (
          <NotificationField
            label={t("detail.notifications.nextAttempt")}
            value={formatDateTime(notification.next_attempt_at)}
          />
        ) : null}
        {notification.last_attempt_at ? (
          <NotificationField
            label={t("detail.notifications.lastAttempt")}
            value={formatDateTime(notification.last_attempt_at)}
          />
        ) : null}
        {notification.delivered_at ? (
          <NotificationField
            label={t("detail.notifications.deliveredAt")}
            value={formatDateTime(notification.delivered_at)}
          />
        ) : null}
        {notification.last_error ? (
          <NotificationField
            label={t("detail.notifications.lastError")}
            value={notification.last_error}
            mono
            className="sm:col-span-2"
          />
        ) : null}
      </dl>
    </li>
  )
}

function NotificationField({
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
    <div className={className ? `min-w-0 space-y-1 ${className}` : "min-w-0 space-y-1"}>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd
        className={`break-all text-sm font-medium ${mono ? "font-mono text-xs" : ""}`}
      >
        {value}
      </dd>
    </div>
  )
}
