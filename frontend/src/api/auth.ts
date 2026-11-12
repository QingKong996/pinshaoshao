import { requestJson } from "@/api/client"

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

export function login(body: LoginRequest) {
  return requestJson<TokenResponse>("/login", {
    method: "POST",
    body: JSON.stringify(body),
  })
}

export function register(body: RegisterRequest) {
  return requestJson<TokenResponse>("/register", {
    method: "POST",
    body: JSON.stringify(body),
  })
}

export function logout(token: string) {
  return requestJson<void>("/logout", {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function merchantLogin(body: LoginRequest) {
  return requestJson<TokenResponse>("/merchant/login", {
    method: "POST",
    body: JSON.stringify(body),
  })
}

export function merchantRegister(body: MerchantRegisterRequest) {
  return requestJson<TokenResponse>("/merchant/register", {
    method: "POST",
    body: JSON.stringify(body),
  })
}

export function merchantLogout(token: string) {
  return requestJson<void>("/merchant/logout", {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  })
}
