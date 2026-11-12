import {
  useMerchantAuth,
  type MerchantSession,
} from "@/auth/merchant/context"

export function useMerchantSession(): MerchantSession {
  const { session } = useMerchantAuth()
  if (!session) {
    throw new Error("useMerchantSession requires an authenticated merchant")
  }
  return session
}
