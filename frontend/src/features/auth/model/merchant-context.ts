import { createContext, useContext } from "react"
import type { MerchantAuthContextValue } from "../types"

export const MerchantAuthContext =
  createContext<MerchantAuthContextValue | null>(null)

export function useMerchantAuth() {
  const context = useContext(MerchantAuthContext)
  if (!context) {
    throw new Error("useMerchantAuth must be used within a MerchantAuthProvider")
  }
  return context
}
