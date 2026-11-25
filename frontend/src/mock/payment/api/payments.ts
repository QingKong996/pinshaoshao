import type {
  CreateMockPaymentInput,
  MockPayment,
  MockPaymentNotification,
  MockPaymentNotificationList,
} from "../dto/payment"
import { mockPayApiClient } from "./client"

const PAYMENTS_PATH = "/payments"

function paymentPath(paymentNo: string) {
  return `${PAYMENTS_PATH}/${encodeURIComponent(paymentNo)}`
}

export async function createMockPayment(input: CreateMockPaymentInput) {
  const { data } = await mockPayApiClient.post<MockPayment>(
    PAYMENTS_PATH,
    input,
  )
  return data
}

export async function getMockPayment(paymentNo: string) {
  const { data } = await mockPayApiClient.get<MockPayment>(
    paymentPath(paymentNo),
  )
  return data
}

export async function closeMockPayment(paymentNo: string) {
  const { data } = await mockPayApiClient.post<MockPayment>(
    `${paymentPath(paymentNo)}/close`,
    undefined,
  )
  return data
}

export async function listMockPaymentNotifications(paymentNo: string) {
  const { data } = await mockPayApiClient.get<MockPaymentNotificationList>(
    `${paymentPath(paymentNo)}/notifications`,
  )
  return data
}

export async function retryMockPaymentNotification(paymentNo: string) {
  const { data } = await mockPayApiClient.post<MockPaymentNotification>(
    `${paymentPath(paymentNo)}/notifications/retry`,
    undefined,
  )
  return data
}
