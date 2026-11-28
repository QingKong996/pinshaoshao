import { render, screen } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { RouterProvider, createMemoryRouter } from "react-router"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { apiClient } from "@/shared/api/client"
import { server } from "@/test/server"

const auth = vi.hoisted(() => ({
  session: null as { token: string; username: string } | null,
  signOut: vi.fn().mockResolvedValue(undefined),
  expireSession: vi.fn((token: string): boolean => auth.session?.token === token),
}))

vi.mock("@/features/auth", () => ({
  useMerchantAuth: () => auth,
  UserMenu: () => <div data-testid="user-menu" />,
}))

import { MerchantConsoleLayout } from "@/app/layouts/merchant-console-layout"

function renderLayout() {
  const router = createMemoryRouter(
    [
      {
        path: "/merchant",
        element: <MerchantConsoleLayout />,
        children: [
          { path: "dashboard", element: <h1>Console home</h1> },
        ],
      },
      {
        path: "/merchant/signin",
        element: <h1>Merchant sign in</h1>,
      },
    ],
    { initialEntries: ["/merchant/dashboard"] },
  )
  return render(<RouterProvider router={router} />)
}

describe("MerchantConsoleLayout", () => {
  beforeEach(() => {
    auth.session = null
    auth.signOut.mockClear()
    auth.expireSession.mockClear()
  })

  it("redirects unauthenticated visitors to merchant sign in", async () => {
    renderLayout()

    expect(await screen.findByText("Merchant sign in")).toBeVisible()
    expect(screen.queryByText("Console home")).not.toBeInTheDocument()
  })

  it("renders the console shell for an active merchant session", () => {
    auth.session = { token: "token", username: "merchant" }
    renderLayout()

    expect(screen.getByText("Console home")).toBeVisible()
    expect(screen.getByTestId("user-menu")).toBeInTheDocument()
  })

  it("clears an expired session and redirects on unauthorized responses", async () => {
    auth.session = { token: "token", username: "merchant" }
    server.use(
      http.get("*/api/v1/expired", () =>
        HttpResponse.json({ error: "expired" }, { status: 401 }),
      ),
    )
    renderLayout()

    await apiClient
      .get("/expired", { headers: { Authorization: "Bearer token" } })
      .catch(() => undefined)

    expect(await screen.findByText("Merchant sign in")).toBeVisible()
    expect(auth.expireSession).toHaveBeenCalledExactlyOnceWith("token")
    expect(auth.signOut).not.toHaveBeenCalled()
  })

  it("ignores an unauthorized response from a previous merchant token", async () => {
    auth.session = { token: "current-token", username: "merchant" }
    server.use(
      http.get("*/api/v1/expired", () =>
        HttpResponse.json({ error: "expired" }, { status: 401 }),
      ),
    )
    renderLayout()

    await apiClient
      .get("/expired", { headers: { Authorization: "Bearer previous-token" } })
      .catch(() => undefined)

    expect(screen.getByText("Console home")).toBeVisible()
    expect(auth.signOut).not.toHaveBeenCalled()
  })
})
