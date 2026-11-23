export const orderQueryKeys = {
  orders: (token: string) => ["orders", token, "list"] as const,
  orderList: (token: string, page: number, pageSize: number) =>
    [...orderQueryKeys.orders(token), page, pageSize] as const,
  order: (token: string, orderID: number) =>
    ["orders", token, "detail", orderID] as const,
}
