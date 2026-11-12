import axios from "axios"

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api/v1"

type ErrorResponse = {
  error: string
}

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
  }
}

type UnauthorizedListener = () => void

const unauthorizedListeners = new Set<UnauthorizedListener>()

// Lets the active session provider clear a stale session when the API rejects its token.
export function onUnauthorized(listener: UnauthorizedListener): () => void {
  unauthorizedListeners.add(listener)
  return () => {
    unauthorizedListeners.delete(listener)
  }
}

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
})

apiClient.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    if (!axios.isAxiosError(error)) {
      return Promise.reject(error)
    }

    const status = error.response?.status ?? 0
    const body = error.response?.data as Partial<ErrorResponse> | undefined
    let message =
      status === 0 ? "network error" : `request failed with status ${status}`
    if (typeof body?.error === "string" && body.error) {
      message = body.error
    }

    if (status === 401) {
      for (const listener of unauthorizedListeners) {
        listener()
      }
    }

    return Promise.reject(new ApiError(status, message))
  },
)
