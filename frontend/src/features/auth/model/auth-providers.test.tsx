import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
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

describe("UserAuthProvider", () => {
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

    const first = render(
      <UserAuthProvider>
        <UserHarness />
      </UserAuthProvider>,
    )
    await user.click(screen.getByRole("button", { name: "Sign in" }))

    expect(await screen.findByLabelText("user session")).toHaveTextContent(
      "shopper",
    )
    expect(localStorage.length).toBe(1)

    first.unmount()
    const restored = render(
      <UserAuthProvider>
        <UserHarness />
      </UserAuthProvider>,
    )
    expect(screen.getByLabelText("user session")).toHaveTextContent("shopper")

    localStorage.setItem(localStorage.key(0)!, "not-json")
    restored.unmount()
    render(
      <UserAuthProvider>
        <UserHarness />
      </UserAuthProvider>,
    )
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

    render(
      <UserAuthProvider>
        <UserHarness />
      </UserAuthProvider>,
    )
    await user.click(screen.getByRole("button", { name: "Sign up" }))
    expect(await screen.findByLabelText("user session")).toHaveTextContent(
      "new-shopper",
    )

    await user.click(screen.getByRole("button", { name: "Sign out" }))

    expect(screen.getByLabelText("user session")).toHaveTextContent(
      "anonymous",
    )
    expect(localStorage.length).toBe(0)
  })
})

describe("MerchantAuthProvider", () => {
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
})
