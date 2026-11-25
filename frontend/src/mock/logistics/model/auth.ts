import { createMockAuth } from "@/mock/auth/provider"
import {
  MOCK_LOGISTICS_API_BASE_URL,
  MOCK_LOGISTICS_SESSION_STORAGE_KEY,
} from "../api/client"

export const mockLogisticsAuth = createMockAuth({
  apiBaseURL: MOCK_LOGISTICS_API_BASE_URL,
  signInPath: "/mocklogistics/signin",
  storageKey: MOCK_LOGISTICS_SESSION_STORAGE_KEY,
})
