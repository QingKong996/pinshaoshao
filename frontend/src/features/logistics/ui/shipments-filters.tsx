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
  SHIPMENT_STATUSES,
  type ShipmentStatusFilter,
} from "../model/filters"

export function ShipmentsFilters({
  statusFilter,
  onStatusFilterChange,
}: {
  statusFilter: ShipmentStatusFilter
  onStatusFilterChange: (status: ShipmentStatusFilter) => void
}) {
  const { t } = useTranslation("logistics")
  const statusItems = [
    { value: ALL, label: t("merchant.filters.allStatuses") },
    ...SHIPMENT_STATUSES.map((status) => ({
      value: status,
      label: t(`status.${status}`),
    })),
  ] satisfies { value: ShipmentStatusFilter; label: string }[]

  return (
    <div className="flex flex-col gap-2 border-b p-4 sm:flex-row sm:items-center">
      <p className="text-sm text-muted-foreground">
        {t("merchant.filters.description")}
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
            className="w-44"
            aria-label={t("merchant.table.status")}
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
