import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { Badge } from "@/shared/ui/badge"
import { formatPriceCents } from "@/shared/lib/format"
import {
  getSKUStatus,
  type ProductSKUStatus,
} from "../model/product"
import {
  buildSKUDimensions,
  isSKUOptionSelectable,
  updateSKUSelection,
  type ProductSKUOption,
  type ProductSKUSelection,
} from "../model/sku-options"
import type { ProductSKU } from "../types"
import { formatProductSpecs } from "./format-specs"

export function ProductSKUPicker({
  skus,
  selectedSKU,
  selectedSpecs,
  onSelectSpecs,
  onSelectSKU,
}: {
  skus: ProductSKU[]
  selectedSKU: ProductSKU | null
  selectedSpecs: ProductSKUSelection
  onSelectSpecs: (specs: ProductSKUSelection) => void
  onSelectSKU: (sku: ProductSKU) => void
}) {
  const { t } = useTranslation("feed")

  if (skus.length === 0) {
    return (
      <p className="rounded-lg border border-dashed px-4 py-6 text-center text-sm text-muted-foreground">
        {t("detail.skus.empty")}
      </p>
    )
  }

  const dimensions = buildSKUDimensions(skus, selectedSpecs)
  if (!dimensions) {
    return (
      <ProductSKUList
        skus={skus}
        selectedSKU={selectedSKU}
        onSelect={onSelectSKU}
      />
    )
  }

  return (
    <div className="space-y-5">
      {dimensions.map((dimension, dimensionIndex) => {
        const priorSelectionComplete = dimensions
          .slice(0, dimensionIndex)
          .every(({ name }) => selectedSpecs[name] !== undefined)

        return (
          <fieldset key={dimension.name} className="min-w-0 space-y-2">
            <legend className="text-sm font-medium">{dimension.name}</legend>
            <div className="flex flex-wrap gap-2">
              {dimension.options.map((option) => {
                const selected =
                  selectedSpecs[dimension.name] === option.value
                const disabled =
                  !priorSelectionComplete ||
                  !isSKUOptionSelectable(option)
                const statusLabel = describeOptionStatus(option, t)

                return (
                  <label
                    key={option.value}
                    className={cn(
                      "flex min-w-24 cursor-pointer flex-col gap-0.5 rounded-lg border bg-background px-3 py-2 transition-colors focus-within:ring-3 focus-within:ring-ring/50",
                      selected
                        ? "border-primary bg-primary/5 ring-1 ring-primary"
                        : "hover:border-primary/40 hover:bg-muted/40",
                      disabled &&
                        "cursor-not-allowed bg-muted/20 opacity-60 hover:border-border hover:bg-muted/20",
                    )}
                  >
                    <input
                      type="radio"
                      name={`sku-${dimension.name}`}
                      aria-label={`${dimension.name}: ${option.value}`}
                      className="sr-only"
                      checked={selected}
                      disabled={disabled}
                      onChange={() =>
                        onSelectSpecs(
                          updateSKUSelection(
                            skus,
                            selectedSpecs,
                            dimension.name,
                            option.value,
                          ),
                        )
                      }
                    />
                    <span className="text-sm font-medium">
                      {option.value}
                    </span>
                    {statusLabel ? (
                      <span
                        className={cn(
                          "text-xs text-muted-foreground",
                          option.status === "out_of_stock" &&
                            "font-medium text-destructive",
                        )}
                      >
                        {statusLabel}
                      </span>
                    ) : null}
                  </label>
                )
              })}
            </div>
          </fieldset>
        )
      })}
    </div>
  )
}

function ProductSKUList({
  skus,
  selectedSKU,
  onSelect,
}: {
  skus: ProductSKU[]
  selectedSKU: ProductSKU | null
  onSelect: (sku: ProductSKU) => void
}) {
  const { t } = useTranslation("feed")

  return (
    <fieldset className="min-w-0">
      <legend className="sr-only">{t("detail.skus.title")}</legend>
      <ul className="grid gap-2">
        {skus.map((sku) => {
          const status = getSKUStatus(sku)
          const selectable = status !== "unavailable"
          const selected = sku.id === selectedSKU?.id
          const statusLabel = describeSKUStatus(status, t)

          return (
            <li key={sku.id}>
              <label
                className={cn(
                  "flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors",
                  selected
                    ? "border-primary bg-primary/5 ring-1 ring-primary"
                    : "hover:border-primary/40 hover:bg-muted/40",
                  selectable
                    ? ""
                    : "cursor-not-allowed bg-muted/20 opacity-60",
                )}
              >
                <input
                  type="radio"
                  name="product-sku"
                  aria-label={formatProductSpecs(
                    sku.specs,
                    t("detail.skus.noSpecs"),
                  )}
                  className="mt-1 size-4 shrink-0 accent-primary disabled:cursor-not-allowed"
                  checked={selected}
                  disabled={!selectable}
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
                    {statusLabel ? (
                      <span
                        className={cn(
                          "mt-1 block text-xs text-muted-foreground",
                          status === "out_of_stock" &&
                            "font-medium text-destructive",
                        )}
                      >
                        {statusLabel}
                      </span>
                    ) : null}
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

function describeOptionStatus(
  option: ProductSKUOption,
  t: TFunction<"feed">,
) {
  if (option.status === "missing") {
    return t("detail.skus.unavailable")
  }
  return describeSKUStatus(option.status, t)
}

function describeSKUStatus(
  status: ProductSKUStatus,
  t: TFunction<"feed">,
) {
  switch (status) {
    case "available":
      return null
    case "out_of_stock":
      return t("detail.skus.outOfStock")
    case "unavailable":
      return t("detail.skus.unavailable")
  }
}
