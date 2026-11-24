export type PaymentStatus = "pending" | "succeeded" | "closed"

export type OrderPayment = {
  payment_no: string
  order_id: number
  provider: string
  provider_transaction_id: string | null
  status: PaymentStatus
  amount_cents: number
  checkout_url?: string
  paid_at: string | null
  created_at: string
  updated_at: string
}

export type StartPaymentResult = {
  payment: OrderPayment
  checkoutURL: string
}
