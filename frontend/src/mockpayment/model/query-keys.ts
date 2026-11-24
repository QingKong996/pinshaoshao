export const mockPaymentQueryKeys = {
  all: ["mock-payments"] as const,
  payment: (paymentNo: string) =>
    [...mockPaymentQueryKeys.all, "payment", paymentNo] as const,
  notifications: (paymentNo: string) =>
    [...mockPaymentQueryKeys.all, "notifications", paymentNo] as const,
}
