import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { Badge } from "@/shared/ui/badge"
import { formatPriceCents } from "@/shared/lib/format"
import { isSKUPurchasable } from "../model/product"
import type { ProductSKU } from "../types"
import { formatProductSpecs } from "./format-specs"

export function ProductSKUPicker({
  skus,
  selectedSKUID,
  onSelect,
}: {
  skus: ProductSKU[]
  selectedSKUID: number | null
  onSelect: (sku: ProductSKU) => void
}) {
  const { t } = useTranslation("feed")

  if (skus.length === 0) {
    return (
      <p className="rounded-lg border border-dashed px-4 py-6 text-center text-sm text-muted-foreground">
        {t("detail.skus.empty")}
      </p>
    )
  }

  return (
    <fieldset className="min-w-0">
      <legend className="sr-only">{t("detail.skus.title")}</legend>
      <ul className="grid gap-2">
        {skus.map((sku) => {
          const purchasable = isSKUPurchasable(sku)
          const selected = sku.id === selectedSKUID

          return (
            <li key={sku.id}>
              <label
                className={cn(
                  "flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors",
                  selected
                    ? "border-primary bg-primary/5 ring-1 ring-primary"
                    : "hover:border-primary/40 hover:bg-muted/40",
                  purchasable
                    ? ""
                    : "cursor-not-allowed bg-muted/20 opacity-60",
                )}
              >
                <input
                  type="radio"
                  name="product-sku"
                  className="mt-1 size-4 shrink-0 accent-primary disabled:cursor-not-allowed"
                  checked={selected}
                  disabled={!purchasable}
                  onChange={() => onSelect(sku)}
                />
                <span className="flex min-w-0 flex-1 flex-wrap items-start justify-between gap-3">
                  <span className="min-w-0">
                    <span className="block font-medium">
                      {formatProductSpecs(
                        sku.specs,
                        t("detail.skus.noSpecs"),
                      )}
                    </span>
                    <span className="mt-1 block text-xs text-muted-foreground">
                      {purchasable
                        ? t("detail.skus.available", {
                            count: sku.available_stock,
                          })
                        : t("detail.skus.unavailable")}
                    </span>
                  </span>
                  <span className="shrink-0 text-right">
                    <span className="block font-heading font-semibold tabular-nums">
                      {formatPriceCents(sku.price_cents)}
                    </span>
                    {selected ? (
                      <Badge variant="secondary" className="mt-1">
                        {t("detail.skus.selected")}
                      </Badge>
                    ) : null}
                  </span>
                </span>
              </label>
            </li>
          )
        })}
      </ul>
    </fieldset>
  )
}
