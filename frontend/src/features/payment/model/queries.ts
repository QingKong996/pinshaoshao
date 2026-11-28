import { useQuery } from "@tanstack/react-query"
import { ApiError } from "@/shared/api/client"
import { getOrderPayment } from "../api/payments"

const REFUND_POLL_INTERVAL_MS = 5_000

export const paymentQueryKeys = {
  orderPayment: (token: string, orderID: number) =>
    ["payment", "order", token, orderID] as const,
}

export function useOrderPayment(token: string, orderID: number) {
  return useQuery({
    queryKey: paymentQueryKeys.orderPayment(token, orderID),
    queryFn: async () => {
      try {
        return await getOrderPayment(token, orderID)
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) {
          return null
        }
        throw error
      }
    },
    enabled: token.length > 0 && Number.isInteger(orderID) && orderID > 0,
    refetchInterval: (query) =>
      query.state.data?.refund?.status === "pending"
        ? REFUND_POLL_INTERVAL_MS
        : false,
  })
}
