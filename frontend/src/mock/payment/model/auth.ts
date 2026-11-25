import { createMockAuth } from "@/mock/auth/provider"
import {
  MOCKPAY_API_BASE_URL,
  MOCKPAY_SESSION_STORAGE_KEY,
} from "../api/client"

export const mockPaymentAuth = createMockAuth({
  apiBaseURL: MOCKPAY_API_BASE_URL,
  signInPath: "/mockpay/signin",
  storageKey: MOCKPAY_SESSION_STORAGE_KEY,
})
