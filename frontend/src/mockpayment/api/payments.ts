import { apiClient } from "@/shared/api/client"
import type {
  CreateMockPaymentInput,
  MockPayment,
  MockPaymentNotification,
  MockPaymentNotificationList,
} from "../dto/payment"

const PAYMENTS_PATH = "/mock-payments"

function paymentPath(paymentNo: string) {
  return `${PAYMENTS_PATH}/${encodeURIComponent(paymentNo)}`
}

export async function createMockPayment(input: CreateMockPaymentInput) {
  const { data } = await apiClient.post<MockPayment>(PAYMENTS_PATH, input)
  return data
}

export async function getMockPayment(paymentNo: string) {
  const { data } = await apiClient.get<MockPayment>(paymentPath(paymentNo))
  return data
}

export async function closeMockPayment(paymentNo: string) {
  const { data } = await apiClient.post<MockPayment>(
    `${paymentPath(paymentNo)}/close`,
  )
  return data
}

export async function listMockPaymentNotifications(paymentNo: string) {
  const { data } = await apiClient.get<MockPaymentNotificationList>(
    `${paymentPath(paymentNo)}/notifications`,
  )
  return data
}

export async function retryMockPaymentNotification(paymentNo: string) {
  const { data } = await apiClient.post<MockPaymentNotification>(
    `${paymentPath(paymentNo)}/notifications/retry`,
  )
  return data
}
