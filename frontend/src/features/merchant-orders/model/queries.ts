import { keepPreviousData, useQuery } from "@tanstack/react-query"
import { getMerchantOrder, listMerchantOrders } from "../api/orders"
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

export function useMerchantOrder(token: string, subOrderID: number) {
  return useQuery({
    queryKey: merchantOrderQueryKeys.order(subOrderID),
    queryFn: () => getMerchantOrder(token, subOrderID),
    enabled: token.length > 0 && Number.isInteger(subOrderID) && subOrderID > 0,
  })
}
