import { keepPreviousData, useQuery } from "@tanstack/react-query"
import { listMerchantOrders } from "../api/orders"
import type { MerchantOrderListQuery } from "../types"
import { merchantOrderQueryKeys } from "./query-keys"

export function useMerchantOrders(
  token: string,
  query: MerchantOrderListQuery,
) {
  return useQuery({
    queryKey: merchantOrderQueryKeys.orderList(query),
    queryFn: () => listMerchantOrders(token, query),
    placeholderData: keepPreviousData,
  })
}
