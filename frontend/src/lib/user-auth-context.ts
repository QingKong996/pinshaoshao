import { createContext, useContext } from "react"
import type { LoginRequest, RegisterRequest } from "@/lib/api"

export type UserSession = {
  token: string
  username: string
}

export type UserAuthContextValue = {
  session: UserSession | null
  signIn: (credentials: LoginRequest) => Promise<void>
  signUp: (details: RegisterRequest) => Promise<void>
  signOut: () => Promise<void>
}

export const UserAuthContext = createContext<UserAuthContextValue | null>(null)

export function useAuth() {
  const context = useContext(UserAuthContext)
  if (!context) {
    throw new Error("useAuth must be used within a UserAuthProvider")
  }
  return context
}
