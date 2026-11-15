import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { MemoryRouter, Route, Routes } from "react-router"
import { describe, expect, it } from "vitest"
import { MerchantAuthProvider } from "@/features/auth/model/merchant-provider"
import { UserAuthProvider } from "@/features/auth/model/user-provider"
import { MerchantSignUpForm } from "@/features/auth/ui/merchant-sign-up-form"
import { SignUpForm } from "@/features/auth/ui/sign-up-form"
import { server } from "@/test/server"

const password = "a".repeat(12)

describe("SignUpForm", () => {
  it("registers without an optional email and redirects", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("*/api/v1/register", async ({ request }) => {
        expect(await request.json()).toEqual({
          username: "new-shopper",
          password,
        })
        return HttpResponse.json({ token: "registered" })
      }),
    )
    render(
      <MemoryRouter initialEntries={["/signup"]}>
        <UserAuthProvider>
          <Routes>
            <Route path="/signup" element={<SignUpForm />} />
            <Route path="/" element={<h1>Account home</h1>} />
          </Routes>
        </UserAuthProvider>
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText("Username"), "new-shopper")
    await user.type(
      screen.getByLabelText("Password", { exact: true }),
      password,
    )
    await user.type(screen.getByLabelText("Confirm password"), password)
    await user.click(screen.getByRole("button", { name: "Sign up" }))

    expect(await screen.findByText("Account home")).toBeVisible()
  })
})

describe("MerchantSignUpForm", () => {
  it("validates confirmation and submits shop registration data", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("*/api/v1/merchant/register", async ({ request }) => {
        expect(await request.json()).toEqual({
          name: "Example Shop",
          username: "merchant",
          email: "owner@example.com",
          password,
        })
        return HttpResponse.json({ token: "merchant-registered" })
      }),
    )
    render(
      <MemoryRouter initialEntries={["/merchant/signup"]}>
        <MerchantAuthProvider onBeforeSessionChange={() => undefined}>
          <Routes>
            <Route
              path="/merchant/signup"
              element={<MerchantSignUpForm redirectTo="/merchant/dashboard" />}
            />
            <Route path="/merchant/dashboard" element={<h1>Merchant home</h1>} />
          </Routes>
        </MerchantAuthProvider>
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText("Shop name"), "Example Shop")
    await user.type(screen.getByLabelText("Shop username"), "merchant")
    await user.type(screen.getByLabelText("Email"), "owner@example.com")
    await user.type(
      screen.getByLabelText("Password", { exact: true }),
      password,
    )
    await user.type(screen.getByLabelText("Confirm password"), "b".repeat(12))
    await user.click(screen.getByRole("button", { name: "Open shop" }))

    expect(await screen.findByText("Passwords do not match")).toBeVisible()

    await user.clear(screen.getByLabelText("Confirm password"))
    await user.type(screen.getByLabelText("Confirm password"), password)
    await user.click(screen.getByRole("button", { name: "Open shop" }))

    expect(await screen.findByText("Merchant home")).toBeVisible()
  })
})
