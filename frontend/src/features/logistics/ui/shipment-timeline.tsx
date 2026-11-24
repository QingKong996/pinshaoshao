import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { formatDateTimeSeconds } from "@/shared/lib/format"
import type { ShipmentEvent } from "../types"

export function ShipmentTimeline({
  events,
  limit,
}: {
  events: ShipmentEvent[]
  limit?: number
}) {
  const { t } = useTranslation("logistics")
  const visibleEvents = [...events]
    .sort((left, right) => right.sequence - left.sequence)
    .slice(0, limit)

  if (visibleEvents.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("timeline.empty")}
      </p>
    )
  }

  return (
    <ol className="space-y-0">
      {visibleEvents.map((event, index) => (
        <li key={event.sequence} className="flex gap-3">
          <div className="flex w-3 shrink-0 flex-col items-center">
            <span
              className={cn(
                "mt-1.5 size-2.5 rounded-full",
                index === 0
                  ? "bg-primary ring-4 ring-primary/15"
                  : "bg-muted-foreground/35",
              )}
              aria-hidden
            />
            {index < visibleEvents.length - 1 ? (
              <span
                className="w-px flex-1 bg-border"
                aria-hidden
              />
            ) : null}
          </div>
          <div className="min-w-0 flex-1 pb-5 last:pb-0">
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
              <p className="font-medium">
                {t(`status.${event.status}`)}
              </p>
              <time className="text-xs text-muted-foreground tabular-nums">
                {formatDateTimeSeconds(event.occurred_at)}
              </time>
            </div>
            <p className="mt-1 text-sm text-muted-foreground">
              {event.description}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {event.location}
            </p>
          </div>
        </li>
      ))}
    </ol>
  )
}
