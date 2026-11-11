import {
  useMerchantAuth,
  type MerchantSession,
} from "@/lib/merchant-auth-context"

export function useMerchantSession(): MerchantSession {
  const { session } = useMerchantAuth()
  if (!session) {
    throw new Error("useMerchantSession requires an authenticated merchant")
  }
  return session
}
