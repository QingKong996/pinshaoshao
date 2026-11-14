import type { ProductListQuery } from "../types"

export const productQueryKeys = {
  products: () => ["merchant", "products"] as const,
  productList: (query: ProductListQuery) =>
    [...productQueryKeys.products(), "list", query] as const,
  product: (productID: number) =>
    [...productQueryKeys.products(), productID] as const,
  productSKUs: (productID: number) =>
    [...productQueryKeys.product(productID), "skus"] as const,
  categories: () => ["merchant", "categories"] as const,
  inventory: () => ["merchant", "inventory"] as const,
}
