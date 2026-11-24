import { useTranslation } from "react-i18next"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/ui/select"
import {
  ALL,
  MERCHANT_ORDER_STATUSES,
  type StatusFilter,
} from "../model/filters"

export function OrdersFilters({
  statusFilter,
  onStatusFilterChange,
}: {
  statusFilter: StatusFilter
  onStatusFilterChange: (status: StatusFilter) => void
}) {
  const { t } = useTranslation("console")
  const statusItems = [
    { value: ALL, label: t("orders.filters.allStatuses") },
    ...MERCHANT_ORDER_STATUSES.map((status) => ({
      value: status,
      label: t(`orders.status.${status}`),
    })),
  ] satisfies { value: StatusFilter; label: string }[]

  return (
    <div className="flex flex-col gap-2 border-b p-4 sm:flex-row sm:items-center">
      <p className="text-sm text-muted-foreground">
        {t("orders.filters.description")}
      </p>
      <div className="flex flex-wrap items-center gap-2 sm:ml-auto">
        <Select
          items={statusItems}
          value={statusFilter}
          onValueChange={(value) => {
            if (value !== null) {
              onStatusFilterChange(value)
            }
          }}
        >
          <SelectTrigger
            className="w-40"
            aria-label={t("orders.table.status")}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {statusItems.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}
