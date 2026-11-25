import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import {
  MOCK_SESSION_REJECTED_EVENT,
  createMockApiClient,
  writeMockSession,
} from "@/shared/api/mock-client"
import { server } from "@/test/server"

describe("createMockApiClient", () => {
  it("uses only its service's sessionStorage token", async () => {
    const storageKey = "pinshaoshao:mockpay-test"
    const client = createMockApiClient("/mockpay/api/v1", storageKey)
    writeMockSession(storageKey, {
      token: "mockpay-token",
      username: "mockpay-operator",
    })
    writeMockSession("pinshaoshao:mocklogistics-test", {
      token: "mocklogistics-token",
      username: "mocklogistics-operator",
    })

    let authorization = ""
    server.use(
      http.get("*/mockpay/api/v1/payments/1", ({ request }) => {
        authorization = request.headers.get("Authorization") ?? ""
        return HttpResponse.json({ ok: true })
      }),
    )

    await client.get("/payments/1")

    expect(authorization).toBe("Bearer mockpay-token")
    expect(
      JSON.parse(
        sessionStorage.getItem("pinshaoshao:mocklogistics-test") ?? "{}",
      ),
    ).toMatchObject({ token: "mocklogistics-token" })
  })

  it("clears only the rejected session and emits its key", async () => {
    const storageKey = "pinshaoshao:mockpay-test"
    const otherStorageKey = "pinshaoshao:mocklogistics-test"
    const client = createMockApiClient("/mockpay/api/v1", storageKey)
    writeMockSession(storageKey, {
      token: "expired-token",
      username: "mockpay-operator",
    })
    writeMockSession(otherStorageKey, {
      token: "logistics-token",
      username: "mocklogistics-operator",
    })
    const listener = vi.fn()
    window.addEventListener(MOCK_SESSION_REJECTED_EVENT, listener)
    server.use(
      http.get("*/mockpay/api/v1/payments/1", () =>
        HttpResponse.json({ error: "invalid session" }, { status: 401 }),
      ),
    )

    await expect(client.get("/payments/1")).rejects.toMatchObject({
      status: 401,
      message: "invalid session",
    })

    expect(sessionStorage.getItem(storageKey)).toBeNull()
    expect(sessionStorage.getItem(otherStorageKey)).not.toBeNull()
    expect(listener).toHaveBeenCalledOnce()
    expect((listener.mock.calls[0][0] as CustomEvent).detail).toBe(storageKey)
    window.removeEventListener(MOCK_SESSION_REJECTED_EVENT, listener)
  })
})
