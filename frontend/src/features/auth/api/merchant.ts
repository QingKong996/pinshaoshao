import { apiClient } from "@/shared/api/client"
import type { LoginRequest, MerchantRegisterRequest, TokenResponse } from "../types"

export async function merchantLogin(body: LoginRequest) {
  const { data } = await apiClient.post<TokenResponse>("/merchant/login", body)
  return data
}

export async function merchantRegister(body: MerchantRegisterRequest) {
  const { data } = await apiClient.post<TokenResponse>("/merchant/register", body)
  return data
}

export async function merchantLogout(token: string) {
  await apiClient.post("/merchant/logout", undefined, {
    headers: { Authorization: `Bearer ${token}` },
  })
}
