export const MOCK_PAYMENT_ROUTE_PATHS = {
  list: "/mockpay",
  detail: "/mockpay/payments/:paymentNo",
} as const

export function mockPaymentDetailPath(paymentNo: string) {
  return `/mockpay/payments/${encodeURIComponent(paymentNo)}`
}
