import { useMutation, useQueryClient } from "@tanstack/react-query"
import { cancelOrder, createDirectOrder, createOrder } from "../api/orders"
import type { CreateDirectOrderInput, CreateOrderInput } from "../types"
import { orderQueryKeys } from "./query-keys"

export function useCreateOrder(token: string, onOrderCreated?: () => void) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateOrderInput) => createOrder(token, input),
    onSuccess: (order) => {
      queryClient.setQueryData(orderQueryKeys.order(token, order.id), order)
      void queryClient.invalidateQueries({
        queryKey: orderQueryKeys.orders(token),
      })
      onOrderCreated?.()
    },
  })
}

export function useCreateDirectOrder(
  token: string,
  onOrderCreated?: () => void,
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateDirectOrderInput) =>
      createDirectOrder(token, input),
    onSuccess: (order) => {
      queryClient.setQueryData(orderQueryKeys.order(token, order.id), order)
      void queryClient.invalidateQueries({
        queryKey: orderQueryKeys.orders(token),
      })
      onOrderCreated?.()
    },
  })
}

export function useCancelOrder(token: string, orderID: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => cancelOrder(token, orderID),
    onSuccess: (order) => {
      queryClient.setQueryData(orderQueryKeys.order(token, orderID), order)
      void queryClient.invalidateQueries({
        queryKey: orderQueryKeys.order(token, orderID),
      })
      void queryClient.invalidateQueries({
        queryKey: orderQueryKeys.orders(token),
      })
    },
  })
}
