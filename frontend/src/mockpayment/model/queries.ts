import { useQuery } from "@tanstack/react-query"
import {
  getMockPayment,
  listMockPaymentNotifications,
} from "../api/payments"
import type { PaymentStatus } from "../dto/payment"
import { mockPaymentQueryKeys } from "./query-keys"

const PAYMENT_POLL_INTERVAL_MS = 3_000
const NOTIFICATION_POLL_INTERVAL_MS = 2_000

export function useMockPayment(paymentNo: string) {
  return useQuery({
    queryKey: mockPaymentQueryKeys.payment(paymentNo),
    queryFn: () => getMockPayment(paymentNo),
    enabled: paymentNo.length > 0,
    refetchInterval: (query) =>
      query.state.data?.status === "pending"
        ? PAYMENT_POLL_INTERVAL_MS
        : false,
  })
}

export function useMockPaymentNotifications(
  paymentNo: string,
  paymentStatus?: PaymentStatus,
) {
  return useQuery({
    queryKey: mockPaymentQueryKeys.notifications(paymentNo),
    queryFn: () => listMockPaymentNotifications(paymentNo),
    enabled: paymentNo.length > 0 && paymentStatus === "paid",
    refetchInterval: (query) => {
      const items = query.state.data?.items
      if (!items) {
        return false
      }
      return items.some(
        (item) => item.status === "pending" || item.status === "delivering",
      )
        ? NOTIFICATION_POLL_INTERVAL_MS
        : false
    },
  })
}
