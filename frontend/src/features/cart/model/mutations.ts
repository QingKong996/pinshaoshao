import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  addCartItem,
  removeCartItem,
  updateCartItemQuantity,
} from "../api/cart"
import type { AddCartItemInput } from "../types"
import { cartQueryKeys } from "./query-keys"

export function useAddCartItem(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: AddCartItemInput) => addCartItem(token, input),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: cartQueryKeys.cart(token),
      }),
  })
}

export function useUpdateCartItemQuantity(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      itemID,
      quantity,
    }: {
      itemID: number
      quantity: number
    }) => updateCartItemQuantity(token, itemID, quantity),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: cartQueryKeys.cart(token),
      }),
  })
}

export function useRemoveCartItem(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (itemID: number) => removeCartItem(token, itemID),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: cartQueryKeys.cart(token),
      }),
  })
}
