import { act, render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { MemoryRouter, Route, Routes } from "react-router"
import { describe, expect, it } from "vitest"
import { useMerchantAuth } from "@/features/auth/model/merchant-context"
import { MerchantAuthProvider } from "@/features/auth/model/merchant-provider"
import { MerchantSignInForm } from "@/features/auth/ui/merchant-sign-in-form"
import { server } from "@/test/server"

const password = "a".repeat(12)
const authToken = ["merchant", "token"].join("-")

function SessionProbe() {
  const { session } = useMerchantAuth()
  return <output aria-label="session">{session?.username ?? "anonymous"}</output>
}

function renderSignInForm() {
  return render(
    <MemoryRouter initialEntries={["/merchant/signin"]}>
      <MerchantAuthProvider onBeforeSessionChange={() => undefined}>
        <SessionProbe />
        <Routes>
          <Route
            path="/merchant/signin"
            element={
              <MerchantSignInForm redirectTo="/merchant/dashboard" />
            }
          />
          <Route path="/merchant/dashboard" element={<h1>Dashboard</h1>} />
        </Routes>
      </MerchantAuthProvider>
    </MemoryRouter>,
  )
}

describe("MerchantSignInForm", () => {
  it("validates credentials before submitting and toggles password visibility", async () => {
    const user = userEvent.setup()
    renderSignInForm()

    await user.click(
      screen.getByRole("button", { name: "Sign in to your shop" }),
    )

    expect(await screen.findByText("Please enter a username")).toBeVisible()
    expect(screen.getByText("Please enter a password")).toBeVisible()

    await user.type(
      screen.getByLabelText("Shop username"),
      "abcd",
    )
    await user.type(screen.getByLabelText("Password"), "short")
    await user.click(
      screen.getByRole("button", { name: "Sign in to your shop" }),
    )

    expect(
      await screen.findByText("Username must be 5-20 characters"),
    ).toBeVisible()
    expect(
      screen.getByText("Password must be 8-20 characters"),
    ).toBeVisible()

    const passwordInput = screen.getByLabelText("Password")
    await user.click(screen.getByRole("button", { name: "Show password" }))
    expect(passwordInput).toHaveAttribute("type", "text")
    await user.click(screen.getByRole("button", { name: "Hide password" }))
    expect(passwordInput).toHaveAttribute("type", "password")
  })

  it("signs in, persists the session, and redirects to the console", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("*/api/v1/merchant/login", async ({ request }) => {
        expect(await request.json()).toEqual({
          username: "merchant",
          password,
        })
        return HttpResponse.json({ token: authToken })
      }),
    )
    renderSignInForm()

    await user.type(screen.getByLabelText("Shop username"), "merchant")
    await user.type(screen.getByLabelText("Password"), password)
    await user.click(
      screen.getByRole("button", { name: "Sign in to your shop" }),
    )

    expect(await screen.findByText("Dashboard")).toBeVisible()
    expect(screen.getByLabelText("session")).toHaveTextContent("merchant")
    expect(localStorage.length).toBe(1)
  })

  it("shows a server error and clears it when credentials change", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("*/api/v1/merchant/login", () =>
        HttpResponse.json({ error: "invalid credentials" }, { status: 401 }),
      ),
    )
    renderSignInForm()

    await user.type(screen.getByLabelText("Shop username"), "merchant")
    await user.type(screen.getByLabelText("Password"), password)
    await user.click(
      screen.getByRole("button", { name: "Sign in to your shop" }),
    )

    expect(
      await screen.findByText("Incorrect username or password"),
    ).toBeVisible()
    await user.type(screen.getByLabelText("Password"), "x")
    expect(
      screen.queryByText("Incorrect username or password"),
    ).not.toBeInTheDocument()
  })

  it("disables submission and announces progress while the request is pending", async () => {
    const user = userEvent.setup()
    let resolveLogin: (response: Response) => void = () => undefined
    const pendingLogin = new Promise<Response>((resolve) => {
      resolveLogin = resolve
    })
    server.use(
      http.post("*/api/v1/merchant/login", () => pendingLogin),
    )
    renderSignInForm()

    await user.type(screen.getByLabelText("Shop username"), "merchant")
    await user.type(screen.getByLabelText("Password"), password)
    await user.click(
      screen.getByRole("button", { name: "Sign in to your shop" }),
    )

    expect(
      screen.getByRole("button", { name: "Signing in..." }),
    ).toBeDisabled()
    expect(
      screen.getByText("Signing in to your shop, please wait"),
    ).toBeInTheDocument()

    await act(async () => {
      resolveLogin(HttpResponse.json({ token: authToken }))
    })
    expect(await screen.findByText("Dashboard")).toBeVisible()
  })
})
