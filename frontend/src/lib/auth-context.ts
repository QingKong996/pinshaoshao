import { createContext, useContext } from "react"
import type {
  LoginRequest,
  MerchantRegisterRequest,
  RegisterRequest,
} from "@/lib/api"

export type AuthKind = "user" | "merchant"

export type AuthSession = {
  token: string
  username: string
}

export type AuthContextValue<
  TRegister extends RegisterRequest = RegisterRequest,
> = {
  session: AuthSession | null
  signIn: (credentials: LoginRequest) => Promise<void>
  signUp: (details: TRegister) => Promise<void>
  signOut: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)
export const MerchantAuthContext = createContext<AuthContextValue<MerchantRegisterRequest> | null>(
  null,
)

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}

// Merchant sessions are kept independent from shopper sessions.
export function useMerchantAuth() {
  const context = useContext(MerchantAuthContext)
  if (!context) {
    throw new Error("useMerchantAuth must be used within a MerchantAuthProvider")
  }
  return context
}
