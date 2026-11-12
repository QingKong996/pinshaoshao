import { apiClient } from "@/shared/api/client"
import type { LoginRequest, RegisterRequest, TokenResponse } from "../types"

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
