import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  closeMockPayment,
  createMockPayment,
  retryMockPaymentNotification,
} from "../api/payments"
import type { MockPaymentNotificationList } from "../dto/payment"
import { mockPaymentQueryKeys } from "./query-keys"

export function useCreateMockPayment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createMockPayment,
    onSuccess: (payment) => {
      queryClient.setQueryData(
        mockPaymentQueryKeys.payment(payment.payment_no),
        payment,
      )
    },
  })
}

export function useCloseMockPayment(paymentNo: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => closeMockPayment(paymentNo),
    onSuccess: (payment) => {
      queryClient.setQueryData(
        mockPaymentQueryKeys.payment(paymentNo),
        payment,
      )
      void queryClient.invalidateQueries({
        queryKey: mockPaymentQueryKeys.notifications(paymentNo),
      })
    },
  })
}

export function useRetryMockPaymentNotification(paymentNo: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => retryMockPaymentNotification(paymentNo),
    onSuccess: (notification) => {
      queryClient.setQueryData<MockPaymentNotificationList>(
        mockPaymentQueryKeys.notifications(paymentNo),
        (current) => {
          if (!current) {
            return current
          }
          return {
            items: current.items.map((item) =>
              item.id === notification.id ? notification : item,
            ),
          }
        },
      )
      void queryClient.invalidateQueries({
        queryKey: mockPaymentQueryKeys.notifications(paymentNo),
      })
    },
  })
}
