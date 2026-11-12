import { createContext, useContext } from "react"
import type { LoginRequest, MerchantRegisterRequest } from "@/api/auth"

export type MerchantSession = {
  token: string
  username: string
}

export type MerchantAuthContextValue = {
  session: MerchantSession | null
  signIn: (credentials: LoginRequest) => Promise<void>
  signUp: (details: MerchantRegisterRequest) => Promise<void>
  signOut: () => Promise<void>
}

export const MerchantAuthContext =
  createContext<MerchantAuthContextValue | null>(null)

export function useMerchantAuth() {
  const context = useContext(MerchantAuthContext)
  if (!context) {
    throw new Error("useMerchantAuth must be used within a MerchantAuthProvider")
  }
  return context
}
