import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { afterEach, describe, expect, it, vi } from "vitest"
import { MerchantAuthProvider } from "@/features/auth/model/merchant-provider"
import { useMerchantAuth } from "@/features/auth/model/merchant-context"
import { UserAuthProvider } from "@/features/auth/model/user-provider"
import { useAuth } from "@/features/auth/model/user-context"
import { server } from "@/test/server"

const password = "a".repeat(12)
const merchantAuthToken = ["merchant", "token"].join("-")

function UserHarness() {
  const { session, signIn, signUp, signOut } = useAuth()

  return (
    <div>
      <output aria-label="user session">
        {session?.username ?? "anonymous"}
      </output>
      <button
        type="button"
        onClick={() =>
          void signIn({ username: "shopper", password })
        }
      >
        Sign in
      </button>
      <button
        type="button"
        onClick={() =>
          void signUp({
            username: "new-shopper",
            password,
            email: "new@example.com",
          })
        }
      >
        Sign up
      </button>
      <button type="button" onClick={() => void signOut()}>
        Sign out
      </button>
    </div>
  )
}

function MerchantHarness() {
  const { session, signIn, signUp, signOut } = useMerchantAuth()

  return (
    <>
      <output aria-label="merchant session">
        {session?.username ?? "anonymous"}
      </output>
      <button
        type="button"
        onClick={() =>
          void signIn({ username: "merchant", password })
        }
      >
        Merchant sign in
      </button>
      <button
        type="button"
        onClick={() =>
          void signUp({
            username: "new-merchant",
            name: "New Shop",
            email: "shop@example.com",
            password,
          })
        }
      >
        Merchant sign up
      </button>
      <button type="button" onClick={() => void signOut()}>
        Merchant sign out
      </button>
    </>
  )
}

function renderMerchantHarness() {
  const onBeforeSessionChange = vi.fn()
  const result = render(
    <MerchantAuthProvider onBeforeSessionChange={onBeforeSessionChange}>
      <MerchantHarness />
    </MerchantAuthProvider>,
  )
  return { ...result, onBeforeSessionChange }
}

function renderUserHarness() {
  const onBeforeSessionChange = vi.fn()
  const result = render(
    <UserAuthProvider onBeforeSessionChange={onBeforeSessionChange}>
      <UserHarness />
    </UserAuthProvider>,
  )
  return { ...result, onBeforeSessionChange }
}

describe("UserAuthProvider", () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("persists sessions, restores valid sessions, and ignores corrupt storage", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("*/api/v1/login", async ({ request }) => {
        expect(await request.json()).toEqual({
          username: "shopper",
          password,
        })
        return HttpResponse.json({ token: "user-token" })
      }),
    )

    const first = renderUserHarness()
    await user.click(screen.getByRole("button", { name: "Sign in" }))

    expect(await screen.findByLabelText("user session")).toHaveTextContent(
      "shopper",
    )
    expect(first.onBeforeSessionChange).toHaveBeenCalledOnce()
    expect(localStorage.length).toBe(1)

    first.unmount()
    const restored = renderUserHarness()
    expect(screen.getByLabelText("user session")).toHaveTextContent("shopper")

    localStorage.setItem(localStorage.key(0)!, "not-json")
    restored.unmount()
    renderUserHarness()
    expect(screen.getByLabelText("user session")).toHaveTextContent(
      "anonymous",
    )
  })

  it("registers a user and clears the session even if logout fails", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("*/api/v1/register", async ({ request }) => {
        expect(await request.json()).toEqual({
          username: "new-shopper",
          password,
          email: "new@example.com",
        })
        return HttpResponse.json({ token: "registered-token" })
      }),
      http.post("*/api/v1/logout", ({ request }) => {
        expect(request.headers.get("Authorization")).toBe(
          "Bearer registered-token",
        )
        return HttpResponse.json({ error: "failed" }, { status: 500 })
      }),
    )

    const rendered = renderUserHarness()
    await user.click(screen.getByRole("button", { name: "Sign up" }))
    expect(await screen.findByLabelText("user session")).toHaveTextContent(
      "new-shopper",
    )

    await user.click(screen.getByRole("button", { name: "Sign out" }))

    expect(screen.getByLabelText("user session")).toHaveTextContent(
      "anonymous",
    )
    expect(rendered.onBeforeSessionChange).toHaveBeenCalledTimes(2)
    expect(localStorage.length).toBe(0)
  })

  it("keeps a user session in memory when storage writes fail", async () => {
    const user = userEvent.setup()
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new DOMException("storage disabled", "SecurityError")
    })
    const removeItem = vi
      .spyOn(Storage.prototype, "removeItem")
      .mockImplementation(() => {
        throw new DOMException("storage disabled", "SecurityError")
      })
    let logoutAuthorization: string | null = null

    server.use(
      http.post("*/api/v1/login", () =>
        HttpResponse.json({ token: "in-memory-user-token" }),
      ),
      http.post("*/api/v1/logout", ({ request }) => {
        logoutAuthorization = request.headers.get("Authorization")
        return HttpResponse.json({})
      }),
    )

    const rendered = renderUserHarness()
    await user.click(screen.getByRole("button", { name: "Sign in" }))

    expect(await screen.findByLabelText("user session")).toHaveTextContent(
      "shopper",
    )
    expect(localStorage.length).toBe(0)

    await user.click(screen.getByRole("button", { name: "Sign out" }))

    expect(screen.getByLabelText("user session")).toHaveTextContent(
      "anonymous",
    )
    expect(removeItem).toHaveBeenCalled()
    expect(rendered.onBeforeSessionChange).toHaveBeenCalledTimes(2)
    await waitFor(() => {
      expect(logoutAuthorization).toBe("Bearer in-memory-user-token")
    })
  })
})

describe("MerchantAuthProvider", () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("restores a merchant session and notifies before session changes", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("*/api/v1/merchant/register", async ({ request }) => {
        expect(await request.json()).toEqual({
          username: "new-merchant",
          name: "New Shop",
          email: "shop@example.com",
          password,
        })
        return HttpResponse.json({ token: merchantAuthToken })
      }),
      http.post("*/api/v1/merchant/logout", ({ request }) => {
        expect(request.headers.get("Authorization")).toBe(
          `Bearer ${merchantAuthToken}`,
        )
        return HttpResponse.json({})
      }),
    )

    const first = renderMerchantHarness()
    await user.click(
      screen.getByRole("button", { name: "Merchant sign up" }),
    )
    expect(await screen.findByLabelText("merchant session")).toHaveTextContent(
      "new-merchant",
    )
    expect(first.onBeforeSessionChange).toHaveBeenCalledOnce()

    first.unmount()
    const restored = renderMerchantHarness()
    expect(screen.getByLabelText("merchant session")).toHaveTextContent(
      "new-merchant",
    )

    await user.click(
      screen.getByRole("button", { name: "Merchant sign out" }),
    )

    expect(screen.getByLabelText("merchant session")).toHaveTextContent(
      "anonymous",
    )
    expect(restored.onBeforeSessionChange).toHaveBeenCalledOnce()
    expect(localStorage.length).toBe(0)
  })

  it("keeps a merchant session in memory when storage writes fail", async () => {
    const user = userEvent.setup()
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new DOMException("storage disabled", "SecurityError")
    })
    const removeItem = vi
      .spyOn(Storage.prototype, "removeItem")
      .mockImplementation(() => {
        throw new DOMException("storage disabled", "SecurityError")
      })
    let logoutAuthorization: string | null = null

    server.use(
      http.post("*/api/v1/merchant/login", () =>
        HttpResponse.json({ token: merchantAuthToken }),
      ),
      http.post("*/api/v1/merchant/logout", ({ request }) => {
        logoutAuthorization = request.headers.get("Authorization")
        return HttpResponse.json({})
      }),
    )

    const rendered = renderMerchantHarness()
    await user.click(
      screen.getByRole("button", { name: "Merchant sign in" }),
    )

    expect(await screen.findByLabelText("merchant session")).toHaveTextContent(
      "merchant",
    )
    expect(localStorage.length).toBe(0)

    await user.click(
      screen.getByRole("button", { name: "Merchant sign out" }),
    )

    expect(screen.getByLabelText("merchant session")).toHaveTextContent(
      "anonymous",
    )
    expect(removeItem).toHaveBeenCalled()
    expect(rendered.onBeforeSessionChange).toHaveBeenCalledTimes(2)
    await waitFor(() => {
      expect(logoutAuthorization).toBe(`Bearer ${merchantAuthToken}`)
    })
  })
})
