import { useQuery } from "@tanstack/react-query"
import { listCart } from "../api/cart"
import { cartQueryKeys } from "./query-keys"

export function useCart(token: string, enabled = true) {
  return useQuery({
    queryKey: cartQueryKeys.cart(token),
    queryFn: () => listCart(token),
    enabled: enabled && token.length > 0,
  })
}
