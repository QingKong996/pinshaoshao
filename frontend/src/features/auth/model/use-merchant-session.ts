import { useMerchantAuth } from "./merchant-context"
import type { MerchantSession } from "../types"

export function useMerchantSession(): MerchantSession {
  const { session } = useMerchantAuth()
  if (!session) {
    throw new Error("useMerchantSession requires an authenticated merchant")
  }
  return session
}
