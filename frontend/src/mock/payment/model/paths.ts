export const MOCK_PAYMENT_ROUTE_PATHS = {
  signIn: "/mockpay/signin",
  list: "/mockpay",
  detail: "/mockpay/payments/:paymentNo",
} as const

export function mockPaymentDetailPath(paymentNo: string) {
  return `/mockpay/payments/${encodeURIComponent(paymentNo)}`
}
