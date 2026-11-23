export const feedQueryKeys = {
  feed: (token: string) => ["feed", token] as const,
  product: (token: string, productID: number) =>
    ["feed", token, "product", productID] as const,
}
