import { createContext, useContext } from "react"
import type { UserAuthContextValue } from "../types"

export const UserAuthContext = createContext<UserAuthContextValue | null>(null)

export function useAuth() {
  const context = useContext(UserAuthContext)
  if (!context) {
    throw new Error("useAuth must be used within a UserAuthProvider")
  }
  return context
}
