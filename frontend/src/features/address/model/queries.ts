import { useQuery } from "@tanstack/react-query"
import { listAddresses } from "../api/addresses"
import { addressQueryKeys } from "./query-keys"

export function useAddresses(token: string) {
  return useQuery({
    queryKey: addressQueryKeys.addresses(token),
    queryFn: () => listAddresses(token),
    enabled: token.length > 0,
  })
}
