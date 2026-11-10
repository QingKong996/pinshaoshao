import { Search } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { Category, MerchantProductStatus } from "@/lib/merchant-api"

export const UNCATEGORIZED_FILTER = "uncategorized"

export type ProductStatusFilter = "" | MerchantProductStatus

type ProductFiltersProps = {
  search: string
  onSearchChange: (value: string) => void
  category: string
  onCategoryChange: (value: string) => void
  status: ProductStatusFilter
  onStatusChange: (value: ProductStatusFilter) => void
  categories: Category[]
}

export function ProductFilters({
  search,
  onSearchChange,
  category,
  onCategoryChange,
  status,
  onStatusChange,
  categories,
}: ProductFiltersProps) {
  const { t } = useTranslation("merchant")

  const categoryLabel =
    category === ""
      ? t("products.filterCategory")
      : category === UNCATEGORIZED_FILTER
        ? t("common.uncategorized")
        : (categories.find((item) => String(item.id) === category)?.name ??
          t("common.uncategorized"))

  const statusLabel =
    status === ""
      ? t("products.filterStatus")
      : status === "on_shelf"
        ? t("status.onShelf")
        : t("status.offShelf")

  return (
    <div className="flex flex-wrap items-center gap-2">
      <div className="relative w-full sm:max-w-xs">
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
      <Select
        value={category}
        onValueChange={(value) => onCategoryChange(value ?? "")}
      >
        <SelectTrigger
          className="w-40"
          aria-label={t("products.categoryFilterLabel")}
        >
          <SelectValue>{categoryLabel}</SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="">{t("common.all")}</SelectItem>
          <SelectItem value={UNCATEGORIZED_FILTER}>
            {t("common.uncategorized")}
          </SelectItem>
          {categories.map((item) => (
            <SelectItem key={item.id} value={String(item.id)}>
              {item.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select
        value={status}
        onValueChange={(value) => onStatusChange(value ?? "")}
      >
        <SelectTrigger
          className="w-32"
          aria-label={t("products.statusFilterLabel")}
        >
          <SelectValue>{statusLabel}</SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="">{t("common.all")}</SelectItem>
          <SelectItem value="on_shelf">{t("status.onShelf")}</SelectItem>
          <SelectItem value="off_shelf">{t("status.offShelf")}</SelectItem>
        </SelectContent>
      </Select>
    </div>
  )
}
