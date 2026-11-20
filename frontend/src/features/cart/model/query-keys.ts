export const cartQueryKeys = {
  cart: (token: string) => ["cart", token] as const,
}
