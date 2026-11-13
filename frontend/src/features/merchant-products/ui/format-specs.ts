import type { ProductSKU } from "../types"

export function formatSpecsLabel(sku: ProductSKU, emptyLabel: string) {
  const entries = Object.entries(sku.specs).sort(([left], [right]) =>
    left.localeCompare(right),
  )
  if (entries.length === 0) {
    return emptyLabel
  }
  return entries.map(([key, value]) => `${key}: ${value}`).join(" / ")
}
