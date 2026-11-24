import { keepPreviousData, useQuery } from "@tanstack/react-query"
import { getOrder, listOrders } from "../api/orders"
import { orderQueryKeys } from "./query-keys"

const PENDING_PAYMENT_POLL_INTERVAL_MS = 3_000

export function useOrders(token: string, page: number, pageSize: number) {
  return useQuery({
    queryKey: orderQueryKeys.orderList(token, page, pageSize),
    queryFn: () => listOrders(token, page, pageSize),
    placeholderData: keepPreviousData,
    enabled: token.length > 0,
  })
}

export function useOrder(token: string, orderID: number) {
  return useQuery({
    queryKey: orderQueryKeys.order(token, orderID),
    queryFn: () => getOrder(token, orderID),
    enabled: token.length > 0 && Number.isInteger(orderID) && orderID > 0,
    refetchInterval: (query) =>
      query.state.data?.status === "pending_payment"
        ? PENDING_PAYMENT_POLL_INTERVAL_MS
        : false,
  })
}
