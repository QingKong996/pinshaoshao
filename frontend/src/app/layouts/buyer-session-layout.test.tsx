import { render, screen } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { RouterProvider, createMemoryRouter } from "react-router"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { BuyerSessionLayout } from "@/app/layouts/buyer-session-layout"
import { apiClient } from "@/shared/api/client"
import { server } from "@/test/server"

const auth = vi.hoisted(() => ({
  session: null as { token: string; username: string } | null,
  signOut: vi.fn().mockResolvedValue(undefined),
  expireSession: vi.fn((token: string): boolean => auth.session?.token === token),
}))

vi.mock("@/features/auth", () => ({
  useAuth: () => auth,
}))

function renderLayout() {
  const router = createMemoryRouter(
    [
      {
        element: <BuyerSessionLayout />,
        children: [{ path: "/orders", element: <h1>Orders</h1> }],
      },
      { path: "/signin", element: <h1>Buyer sign in</h1> },
    ],
    { initialEntries: ["/orders"] },
  )
  return render(<RouterProvider router={router} />)
}

describe("BuyerSessionLayout", () => {
  beforeEach(() => {
    auth.session = null
    auth.signOut.mockClear()
    auth.expireSession.mockClear()
  })

  it("renders buyer routes without requiring an active session", () => {
    renderLayout()

    expect(screen.getByText("Orders")).toBeVisible()
  })

  it("clears an expired buyer session and redirects on unauthorized responses", async () => {
    auth.session = { token: "token", username: "buyer" }
    server.use(
      http.get("*/api/v1/expired", () =>
        HttpResponse.json({ error: "expired" }, { status: 401 }),
      ),
    )
    renderLayout()

    await apiClient
      .get("/expired", { headers: { Authorization: "Bearer token" } })
      .catch(() => undefined)

    expect(await screen.findByText("Buyer sign in")).toBeVisible()
    expect(auth.expireSession).toHaveBeenCalledExactlyOnceWith("token")
    expect(auth.signOut).not.toHaveBeenCalled()
  })

  it("ignores an unauthorized response from a previous buyer token", async () => {
    auth.session = { token: "current-token", username: "buyer" }
    server.use(
      http.get("*/api/v1/expired", () =>
        HttpResponse.json({ error: "expired" }, { status: 401 }),
      ),
    )
    renderLayout()

    await apiClient
      .get("/expired", { headers: { Authorization: "Bearer previous-token" } })
      .catch(() => undefined)

    expect(screen.getByText("Orders")).toBeVisible()
    expect(auth.signOut).not.toHaveBeenCalled()
  })
})
