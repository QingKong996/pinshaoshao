import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type {
  CreateDirectOrderInput,
  CreateOrderInput,
  Order,
  OrderListResponse,
} from "../types"

export async function createOrder(token: string, input: CreateOrderInput) {
  const { data } = await apiClient.post<Order>(
    "/orders",
    input,
    authorized(token),
  )
  return data
}

export async function createDirectOrder(
  token: string,
  input: CreateDirectOrderInput,
) {
  const { data } = await apiClient.post<Order>(
    "/orders/direct",
    input,
    authorized(token),
  )
  return data
}

export async function listOrders(token: string, page: number, pageSize: number) {
  const { data } = await apiClient.get<OrderListResponse>(
    "/orders",
    {
      ...authorized(token),
      params: { page, page_size: pageSize },
    },
  )
  return data
}

export async function getOrder(token: string, orderID: number) {
  const { data } = await apiClient.get<Order>(
    `/orders/${orderID}`,
    authorized(token),
  )
  return data
}

export async function cancelOrder(token: string, orderID: number) {
  const { data } = await apiClient.post<Order>(
    `/orders/${orderID}/cancel`,
    undefined,
    authorized(token),
  )
  return data
}
