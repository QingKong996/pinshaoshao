import { createContext, useContext } from "react"
import type { LoginRequest, RegisterRequest } from "@/lib/api"

export type AuthSession = {
  token: string
  username: string
}

export type AuthContextValue = {
  session: AuthSession | null
  signIn: (credentials: LoginRequest) => Promise<void>
  signUp: (details: RegisterRequest) => Promise<void>
  signOut: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
