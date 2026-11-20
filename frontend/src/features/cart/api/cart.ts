import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { AddCartItemInput, CartItem } from "../types"

export async function listCart(token: string) {
  const { data } = await apiClient.get<CartItem[]>("/cart", authorized(token))
  return data
}

export async function addCartItem(
  token: string,
  input: AddCartItemInput,
) {
  const { data } = await apiClient.post<CartItem>(
    "/cart/items",
    input,
    authorized(token),
  )
  return data
}

export async function updateCartItemQuantity(
  token: string,
  itemID: number,
  quantity: number,
) {
  const { data } = await apiClient.patch<CartItem>(
    `/cart/items/${itemID}`,
    { quantity },
    authorized(token),
  )
  return data
}

export async function removeCartItem(token: string, itemID: number) {
  await apiClient.delete(`/cart/items/${itemID}`, authorized(token))
}
