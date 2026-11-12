export const productQueryKeys = {
  products: () => ["merchant", "products"] as const,
  product: (productID: number) =>
    [...productQueryKeys.products(), productID] as const,
  productSKUs: (productID: number) =>
    [...productQueryKeys.product(productID), "skus"] as const,
  categories: () => ["merchant", "categories"] as const,
}
