import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import {
  ApiError,
  apiClient,
  onUnauthorized,
} from "@/shared/api/client"
import { server } from "@/test/server"

describe("apiClient", () => {
  it("preserves the server error message and status", async () => {
    server.use(
      http.get("*/api/v1/error", () =>
        HttpResponse.json(
          { error: "category name already exists" },
          { status: 409 },
        ),
      ),
    )

    const error = await apiClient.get("/error").catch((reason) => reason)

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      name: "ApiError",
      status: 409,
      message: "category name already exists",
    })
  })

  it("uses a network error status when no response is received", async () => {
    server.use(http.get("*/api/v1/offline", () => HttpResponse.error()))

    await expect(apiClient.get("/offline")).rejects.toMatchObject({
      name: "ApiError",
      status: 0,
      message: "network error",
    })
  })

  it("notifies and unsubscribes unauthorized listeners", async () => {
    const listener = vi.fn()
    const unsubscribe = onUnauthorized(listener)
    server.use(
      http.get("*/api/v1/protected", () =>
        HttpResponse.json({ error: "expired" }, { status: 401 }),
      ),
    )

    await expect(
      apiClient.get("/protected", {
        headers: { Authorization: "Bearer current-token" },
      }),
    ).rejects.toMatchObject({ status: 401, message: "expired" })
    expect(listener).toHaveBeenCalledExactlyOnceWith("current-token")

    unsubscribe()
    await expect(
      apiClient.get("/protected", {
        headers: { Authorization: "Bearer current-token" },
      }),
    ).rejects.toMatchObject({ status: 401 })
    expect(listener).toHaveBeenCalledExactlyOnceWith("current-token")
  })

  it("does not notify session listeners for unauthenticated 401 responses", async () => {
    const listener = vi.fn()
    const unsubscribe = onUnauthorized(listener)
    server.use(
      http.post("*/api/v1/login", () =>
        HttpResponse.json({ error: "invalid credentials" }, { status: 401 }),
      ),
    )

    await expect(apiClient.post("/login", {})).rejects.toMatchObject({
      status: 401,
    })

    expect(listener).not.toHaveBeenCalled()
    unsubscribe()
  })

  it("falls back to the status when the response body has no error", async () => {
    server.use(
      http.get("*/api/v1/unavailable", () =>
        new HttpResponse(null, { status: 503 }),
      ),
    )

    await expect(apiClient.get("/unavailable")).rejects.toMatchObject({
      status: 503,
      message: "request failed with status 503",
    })
  })
})
