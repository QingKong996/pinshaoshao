import { useTranslation } from "react-i18next"
import { Card, CardContent } from "@/shared/ui/card"
import { formatDateTime, formatPriceRange } from "@/shared/lib/format"
import type { Product } from "../types"

export function ProductMeta({ product }: { product: Product }) {
  const { t } = useTranslation("console")
  const priceRange = formatPriceRange(
    product.min_price_cents,
    product.max_price_cents,
  )
  const items = [
    {
      label: t("productDetail.meta.category"),
      value: product.category_name ?? t("products.filters.uncategorized"),
    },
    {
      label: t("productDetail.meta.skuCount"),
      value: String(product.sku_count),
    },
    {
      label: t("productDetail.meta.stock"),
      value: String(product.total_stock),
    },
    { label: t("productDetail.meta.price"), value: priceRange ?? "—" },
    {
      label: t("productDetail.meta.updatedAt"),
      value: formatDateTime(product.updated_at),
    },
  ]

  return (
    <Card>
      <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        {items.map((item) => (
          <div key={item.label} className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{item.label}</span>
            <span className="font-medium tabular-nums">{item.value}</span>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
