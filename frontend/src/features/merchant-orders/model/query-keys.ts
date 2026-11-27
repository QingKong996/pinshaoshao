import type { MerchantOrderListQuery } from "../types"

export const merchantOrderQueryKeys = {
  orders: () => ["merchant", "orders"] as const,
  orderList: (query: MerchantOrderListQuery) =>
    [...merchantOrderQueryKeys.orders(), "list", query] as const,
  order: (subOrderID: number) =>
    [...merchantOrderQueryKeys.orders(), "detail", subOrderID] as const,
  receiver: (subOrderID: number) =>
    [...merchantOrderQueryKeys.order(subOrderID), "receiver"] as const,
}
