import { useMutation } from "@tanstack/react-query"
import {
  createOrderPayment,
  payOrderPayment,
} from "../api/payments"
import type { StartPaymentResult } from "../types"

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
