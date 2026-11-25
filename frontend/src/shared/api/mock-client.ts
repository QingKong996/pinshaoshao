import axios from "axios"
import { ApiError } from "./client"

type ErrorResponse = {
  error: string
}

export type MockSession = {
  token: string
  username: string
}

export const MOCK_SESSION_REJECTED_EVENT = "mock-session-rejected"

export function readMockSession(storageKey: string): MockSession | null {
  try {
    const raw = sessionStorage.getItem(storageKey)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<MockSession>
    if (
      typeof parsed.token !== "string" ||
      !parsed.token ||
      typeof parsed.username !== "string" ||
      !parsed.username
    ) {
      return null
    }
    return { token: parsed.token, username: parsed.username }
  } catch {
    return null
  }
}

export function writeMockSession(
  storageKey: string,
  session: MockSession | null,
) {
  if (session) {
    sessionStorage.setItem(storageKey, JSON.stringify(session))
    return
  }
  sessionStorage.removeItem(storageKey)
}

export function createMockApiClient(baseURL: string, storageKey: string) {
  const client = axios.create({
    baseURL,
    headers: {
      "Content-Type": "application/json",
    },
  })

  client.interceptors.request.use((config) => {
    const token = readMockSession(storageKey)?.token
    if (token && !config.headers.Authorization) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })

  client.interceptors.response.use(
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
        writeMockSession(storageKey, null)
        window.dispatchEvent(
          new CustomEvent(MOCK_SESSION_REJECTED_EVENT, {
            detail: storageKey,
          }),
        )
      }

      return Promise.reject(new ApiError(status, message))
    },
  )

  return client
}
