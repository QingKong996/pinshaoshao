import { createMockApiClient } from "@/shared/api/mock-client"

export const MOCKPAY_API_BASE_URL = "/mockpay/api/v1"
export const MOCKPAY_SESSION_STORAGE_KEY = "pinshaoshao:mockpay-session"

export const mockPayApiClient = createMockApiClient(
  MOCKPAY_API_BASE_URL,
  MOCKPAY_SESSION_STORAGE_KEY,
)
