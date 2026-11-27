import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { Address, AddressInput } from "../types"

export async function listAddresses(token: string) {
  const { data } = await apiClient.get<Address[]>(
    "/addresses",
    authorized(token),
  )
  return data
}

export async function createAddress(token: string, input: AddressInput) {
  const { data } = await apiClient.post<Address>(
    "/addresses",
    input,
    authorized(token),
  )
  return data
}

export async function updateAddress(
  token: string,
  addressID: number,
  input: AddressInput,
) {
  const { data } = await apiClient.put<Address>(
    `/addresses/${addressID}`,
    input,
    authorized(token),
  )
  return data
}

export async function deleteAddress(token: string, addressID: number) {
  await apiClient.delete(`/addresses/${addressID}`, authorized(token))
}

export async function setDefaultAddress(token: string, addressID: number) {
  const { data } = await apiClient.post<Address>(
    `/addresses/${addressID}/default`,
    undefined,
    authorized(token),
  )
  return data
}
