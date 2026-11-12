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

export async function requestJson<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    })
  } catch {
    throw new ApiError(0, "network error")
  }

  if (!response.ok) {
    let message = `request failed with status ${response.status}`
    try {
      const body = (await response.json()) as Partial<ErrorResponse>
      if (typeof body.error === "string" && body.error) {
        message = body.error
      }
    } catch {
      // keep the status based message
    }
    if (response.status === 401) {
      for (const listener of unauthorizedListeners) {
        listener()
      }
    }
    throw new ApiError(response.status, message)
  }

  if (response.status === 204) {
    return undefined as T
  }
  return (await response.json()) as T
}
