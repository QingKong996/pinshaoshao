import { authorized } from "@/shared/api/authorized"
import { apiClient } from "@/shared/api/client"
import type { OrderPayment } from "../types"

const PAYMENTS_PATH = "/payments"

function paymentPath(paymentNo: string) {
  return `${PAYMENTS_PATH}/${encodeURIComponent(paymentNo)}`
}

export async function createOrderPayment(token: string, orderID: number) {
  const { data } = await apiClient.post<OrderPayment>(
    PAYMENTS_PATH,
    { order_id: orderID },
    authorized(token),
  )
  return data
}

export async function getOrderPayment(token: string, orderID: number) {
  const { data } = await apiClient.get<OrderPayment>(
    `/orders/${orderID}/payment`,
    authorized(token),
  )
  return data
}

export async function payOrderPayment(token: string, paymentNo: string) {
  const { data } = await apiClient.post<OrderPayment>(
    `${paymentPath(paymentNo)}/pay`,
    undefined,
    authorized(token),
  )
  return data
}

export async function closeOrderPayment(token: string, paymentNo: string) {
  const { data } = await apiClient.post<OrderPayment>(
    `${paymentPath(paymentNo)}/close`,
    undefined,
    authorized(token),
  )
  return data
}
