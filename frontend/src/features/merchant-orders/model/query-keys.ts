import type { MerchantOrderListQuery } from "../types"

export const merchantOrderQueryKeys = {
  orders: () => ["merchant", "orders"] as const,
  orderList: (query: MerchantOrderListQuery) =>
    [...merchantOrderQueryKeys.orders(), "list", query] as const,
}
