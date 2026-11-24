export const MOCK_PAYMENT_ROUTE_PATHS = {
  list: "/mock-payments",
  detail: "/mock-payments/:paymentNo",
} as const

export function mockPaymentDetailPath(paymentNo: string) {
  return `/mock-payments/${encodeURIComponent(paymentNo)}`
}
