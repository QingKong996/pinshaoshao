import type { ProductSKU } from "../types"

export function isSKUPurchasable(sku: ProductSKU) {
  return sku.is_enabled && sku.available_stock > 0
}
