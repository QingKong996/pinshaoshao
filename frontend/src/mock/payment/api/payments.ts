import { apiClient } from "@/shared/api/client"
import type {
  CreateMockPaymentInput,
  MockPayment,
  MockPaymentNotification,
  MockPaymentNotificationList,
} from "../dto/payment"

const MOCKPAY_API_BASE_URL = "/mockpay/api/v1"
const PAYMENTS_PATH = "/payments"

function paymentPath(paymentNo: string) {
  return `${PAYMENTS_PATH}/${encodeURIComponent(paymentNo)}`
}

export async function createMockPayment(input: CreateMockPaymentInput) {
  const { data } = await apiClient.post<MockPayment>(PAYMENTS_PATH, input, {
    baseURL: MOCKPAY_API_BASE_URL,
  })
  return data
}

export async function getMockPayment(paymentNo: string) {
  const { data } = await apiClient.get<MockPayment>(paymentPath(paymentNo), {
    baseURL: MOCKPAY_API_BASE_URL,
  })
  return data
}

export async function closeMockPayment(paymentNo: string) {
  const { data } = await apiClient.post<MockPayment>(
    `${paymentPath(paymentNo)}/close`,
    undefined,
    { baseURL: MOCKPAY_API_BASE_URL },
  )
  return data
}

export async function listMockPaymentNotifications(paymentNo: string) {
  const { data } = await apiClient.get<MockPaymentNotificationList>(
    `${paymentPath(paymentNo)}/notifications`,
    { baseURL: MOCKPAY_API_BASE_URL },
  )
  return data
}

export async function retryMockPaymentNotification(paymentNo: string) {
  const { data } = await apiClient.post<MockPaymentNotification>(
    `${paymentPath(paymentNo)}/notifications/retry`,
    undefined,
    { baseURL: MOCKPAY_API_BASE_URL },
  )
  return data
}
