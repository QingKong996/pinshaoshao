import { createMockApiClient } from "@/shared/api/mock-client"

export const MOCK_LOGISTICS_API_BASE_URL = "/mocklogistics/api/v1"
export const MOCK_LOGISTICS_SESSION_STORAGE_KEY =
  "pinshaoshao:mocklogistics-session"

export const mockLogisticsApiClient = createMockApiClient(
  MOCK_LOGISTICS_API_BASE_URL,
  MOCK_LOGISTICS_SESSION_STORAGE_KEY,
)
