import { useMutation } from "@tanstack/react-query"
import {
  closeOrderPayment,
  createOrderPayment,
  payOrderPayment,
} from "../api/payments"
import type { OrderPayment, StartPaymentResult } from "../types"

export function useStartPayment(token: string) {
  return useMutation({
    mutationFn: async (orderID: number): Promise<StartPaymentResult> => {
      const payment = await createOrderPayment(token, orderID)
      if (payment.status !== "pending") {
        return { payment, checkoutURL: "" }
      }

      const session = await payOrderPayment(token, payment.payment_no)
      return {
        payment: session,
        checkoutURL: session.checkout_url?.trim() ?? "",
      }
    },
  })
}

export function useClosePayment(token: string) {
  return useMutation({
    mutationFn: async (orderID: number): Promise<OrderPayment> => {
      const payment = await createOrderPayment(token, orderID)
      if (payment.status !== "pending") {
        return payment
      }
      return closeOrderPayment(token, payment.payment_no)
    },
  })
}
