import { Search } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Input } from "@/shared/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/ui/select"
import { ALL, UNCATEGORIZED, type StatusFilter } from "../model/filters"
import type { ProductCategory } from "../types"

export function ProductsFilters({
  search,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  categoryFilter,
  onCategoryFilterChange,
  categories,
}: {
  search: string
  onSearchChange: (search: string) => void
  statusFilter: StatusFilter
  onStatusFilterChange: (status: StatusFilter) => void
  categoryFilter: string
  onCategoryFilterChange: (category: string) => void
  categories: ProductCategory[]
}) {
  const { t } = useTranslation("console")
  const statusItems = [
    { value: ALL, label: t("products.filters.allStatuses") },
    { value: "on_shelf", label: t("productStatus.on_shelf") },
    { value: "off_shelf", label: t("productStatus.off_shelf") },
  ] satisfies { value: StatusFilter; label: string }[]
  const categoryItems = [
    { value: ALL, label: t("products.filters.allCategories") },
    { value: UNCATEGORIZED, label: t("products.filters.uncategorized") },
    ...categories.map((category) => ({
      value: String(category.id),
      label: category.name,
    })),
  ]

  return (
    <div className="flex flex-col gap-2 border-b p-4 lg:flex-row lg:items-center">
      <div className="relative w-full lg:max-w-xs">
        <Search
          className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground"
          aria-hidden
        />
        <Input
          value={search}
          onChange={(event) => onSearchChange(event.target.value)}
          placeholder={t("products.searchPlaceholder")}
          aria-label={t("products.searchPlaceholder")}
          className="pl-8"
        />
      </div>
      <div className="flex flex-wrap items-center gap-2 lg:ml-auto">
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
            className="w-36"
            aria-label={t("products.table.status")}
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
        <Select
          items={categoryItems}
          value={categoryFilter}
          onValueChange={(value) => {
            if (value !== null) {
              onCategoryFilterChange(value)
            }
          }}
        >
          <SelectTrigger
            className="w-44"
            aria-label={t("products.table.category")}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {categoryItems.map((item) => (
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
