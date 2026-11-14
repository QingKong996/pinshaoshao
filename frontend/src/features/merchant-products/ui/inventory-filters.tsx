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
import { ALL, type StockFilter } from "../model/inventory"

type CategoryFilterItem = {
  value: string
  label: string
}

export function InventoryFilters({
  search,
  onSearchChange,
  categoryItems,
  categoryFilter,
  onCategoryFilterChange,
  stockFilter,
  onStockFilterChange,
}: {
  search: string
  onSearchChange: (search: string) => void
  categoryItems: CategoryFilterItem[]
  categoryFilter: string
  onCategoryFilterChange: (category: string) => void
  stockFilter: StockFilter
  onStockFilterChange: (stock: StockFilter) => void
}) {
  const { t } = useTranslation("console")
  const stockItems = [
    { value: ALL, label: t("inventory.filters.allStock") },
    { value: "in_stock", label: t("inventory.filters.inStock") },
    { value: "out_of_stock", label: t("inventory.filters.outOfStock") },
  ] satisfies { value: StockFilter; label: string }[]
  const categoryOptions = [
    { value: ALL, label: t("inventory.filters.allCategories") },
    ...categoryItems,
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
          placeholder={t("inventory.searchPlaceholder")}
          aria-label={t("inventory.searchPlaceholder")}
          className="pl-8"
        />
      </div>
      <div className="flex flex-wrap items-center gap-2 lg:ml-auto">
        <Select
          items={categoryOptions}
          value={categoryFilter}
          onValueChange={(value) => {
            if (value !== null) {
              onCategoryFilterChange(value)
            }
          }}
        >
          <SelectTrigger
            className="w-44"
            aria-label={t("inventory.table.category")}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {categoryOptions.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select
          items={stockItems}
          value={stockFilter}
          onValueChange={(value) => {
            if (value !== null) {
              onStockFilterChange(value)
            }
          }}
        >
          <SelectTrigger
            className="w-36"
            aria-label={t("inventory.table.stock")}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {stockItems.map((item) => (
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
