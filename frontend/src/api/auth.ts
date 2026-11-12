import { apiClient } from "@/api/client"

export type LoginRequest = {
  username: string
  password: string
}

export type RegisterRequest = {
  username: string
  password: string
  email?: string
}

export type MerchantRegisterRequest = {
  username: string
  name: string
  password: string
  email: string
}

export type TokenResponse = {
  token: string
}

export async function login(body: LoginRequest) {
  const { data } = await apiClient.post<TokenResponse>("/login", body)
  return data
}

export async function register(body: RegisterRequest) {
  const { data } = await apiClient.post<TokenResponse>("/register", body)
  return data
}

export async function logout(token: string) {
  await apiClient.post("/logout", undefined, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

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
