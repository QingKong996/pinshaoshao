export const merchantConsoleQueryKeys = {
  all: ["merchant"] as const,
  dashboard: () => [...merchantConsoleQueryKeys.all, "dashboard"] as const,
  shop: () => [...merchantConsoleQueryKeys.all, "shop"] as const,
  products: () => [...merchantConsoleQueryKeys.all, "products"] as const,
  product: (productID: number) =>
    [...merchantConsoleQueryKeys.products(), productID] as const,
  productSKUs: (productID: number) =>
    [...merchantConsoleQueryKeys.product(productID), "skus"] as const,
  categories: () => [...merchantConsoleQueryKeys.all, "categories"] as const,
}
