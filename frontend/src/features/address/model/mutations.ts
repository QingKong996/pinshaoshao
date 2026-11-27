import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  createAddress,
  deleteAddress,
  setDefaultAddress,
  updateAddress,
} from "../api/addresses"
import type { AddressInput } from "../types"
import { addressQueryKeys } from "./query-keys"

type UpdateAddressVariables = {
  addressID: number
  input: AddressInput
}

function useAddressMutationInvalidation(token: string) {
  const queryClient = useQueryClient()
  return () =>
    queryClient.invalidateQueries({
      queryKey: addressQueryKeys.addresses(token),
    })
}

export function useCreateAddress(token: string) {
  const invalidate = useAddressMutationInvalidation(token)
  return useMutation({
    mutationFn: (input: AddressInput) => createAddress(token, input),
    onSuccess: invalidate,
  })
}

export function useUpdateAddress(token: string) {
  const invalidate = useAddressMutationInvalidation(token)
  return useMutation({
    mutationFn: ({ addressID, input }: UpdateAddressVariables) =>
      updateAddress(token, addressID, input),
    onSuccess: invalidate,
  })
}

export function useDeleteAddress(token: string) {
  const invalidate = useAddressMutationInvalidation(token)
  return useMutation({
    mutationFn: (addressID: number) => deleteAddress(token, addressID),
    onSuccess: invalidate,
  })
}

export function useSetDefaultAddress(token: string) {
  const invalidate = useAddressMutationInvalidation(token)
  return useMutation({
    mutationFn: (addressID: number) => setDefaultAddress(token, addressID),
    onSuccess: invalidate,
  })
}
