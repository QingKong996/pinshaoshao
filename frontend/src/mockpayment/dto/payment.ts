export type PaymentStatus = "pending" | "paid" | "closed"

export type NotificationStatus =
  | "pending"
  | "delivering"
  | "succeeded"
  | "exhausted"

export type MockPayment = {
  payment_no: string
  merchant_order_no: string
  status: PaymentStatus
  amount_cents: number
  currency: string
  subject: string
  checkout_url: string
  return_url?: string
  expires_at: string
  paid_at: string | null
  closed_at: string | null
  created_at: string
}

export type MockPaymentNotification = {
  id: number
  event_id: string
  event_type: string
  target_url: string
  status: NotificationStatus
  attempt_count: number
  max_attempts: number
  next_attempt_at: string
  last_attempt_at: string | null
  delivered_at: string | null
  last_error: string | null
  created_at: string
  updated_at: string
}

export type MockPaymentNotificationList = {
  items: MockPaymentNotification[]
}

export type CreateMockPaymentInput = {
  merchant_order_no: string
  amount_cents: number
  currency: string
  subject: string
  notify_url: string
  return_url?: string
  expires_in_seconds: number
}
