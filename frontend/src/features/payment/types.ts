export type PaymentStatus = "pending" | "succeeded" | "closed"

export type RefundStatus = "pending" | "succeeded" | "failed"

export type PaymentRefund = {
  refund_no: string
  amount_cents: number
  reason: string
  status: RefundStatus
  provider_refund_no: string | null
  refunded_at: string | null
  failed_at: string | null
  created_at: string
  updated_at: string
}

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
  refund?: PaymentRefund
}

export type StartPaymentResult = {
  payment: OrderPayment
  checkoutURL: string
}
