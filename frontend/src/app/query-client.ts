import { QueryClient } from "@tanstack/react-query"
import { ApiError } from "@/shared/api/client"

const STALE_TIME_MS = 30_000
const MAX_RETRIES = 2

function shouldRetry(failureCount: number, error: unknown) {
  if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
    return false
  }
  return failureCount < MAX_RETRIES
}

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: STALE_TIME_MS,
      retry: shouldRetry,
    },
  },
})
