import type { ProductSKU } from "../types"

export type ProductSKUStatus = "available" | "out_of_stock" | "unavailable"

export function isSKUPurchasable(sku: ProductSKU) {
  return sku.is_enabled && sku.in_stock
}

export function getSKUStatus(sku: ProductSKU): ProductSKUStatus {
  if (isSKUPurchasable(sku)) {
    return "available"
  }
  return sku.is_enabled ? "out_of_stock" : "unavailable"
}

export function findPreferredSKU(skus: ProductSKU[]) {
  return (
    skus.find(isSKUPurchasable) ??
    skus.find((sku) => sku.is_enabled) ??
    skus[0] ??
    null
  )
}
