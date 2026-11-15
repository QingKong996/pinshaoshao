import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { MemoryRouter, Route, Routes } from "react-router"
import { describe, expect, it } from "vitest"
import { useAuth } from "@/features/auth/model/user-context"
import { UserAuthProvider } from "@/features/auth/model/user-provider"
import { SignInForm } from "@/features/auth/ui/sign-in-form"
import { server } from "@/test/server"

const password = "a".repeat(12)

function SessionProbe() {
  const { session } = useAuth()
  return <output aria-label="session">{session?.username ?? "anonymous"}</output>
}

function renderSignInForm() {
  return render(
    <MemoryRouter initialEntries={["/signin"]}>
      <UserAuthProvider>
        <SessionProbe />
        <Routes>
          <Route path="/signin" element={<SignInForm />} />
          <Route path="/" element={<h1>Account home</h1>} />
        </Routes>
      </UserAuthProvider>
    </MemoryRouter>,
  )
}

describe("SignInForm", () => {
  it("validates required credentials before submitting", async () => {
    const user = userEvent.setup()
    renderSignInForm()

    await user.click(screen.getByRole("button", { name: "Sign in" }))

    expect(await screen.findByText("Please enter a username")).toBeVisible()
    expect(screen.getByText("Please enter a password")).toBeVisible()
  })

  it("signs in, persists the session, and redirects home", async () => {
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
    renderSignInForm()

    await user.type(screen.getByLabelText("Username"), "shopper")
    await user.type(screen.getByLabelText("Password"), password)
    await user.click(screen.getByRole("button", { name: "Sign in" }))

    expect(await screen.findByText("Account home")).toBeVisible()
    expect(screen.getByLabelText("session")).toHaveTextContent("shopper")
    expect(localStorage.length).toBe(1)
  })

  it("shows a login error and clears it when credentials change", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("*/api/v1/login", () =>
        HttpResponse.json({ error: "invalid credentials" }, { status: 401 }),
      ),
    )
    renderSignInForm()

    await user.type(screen.getByLabelText("Username"), "shopper")
    await user.type(screen.getByLabelText("Password"), password)
    await user.click(screen.getByRole("button", { name: "Sign in" }))

    expect(
      await screen.findByText("Incorrect username or password"),
    ).toBeVisible()
    await user.type(screen.getByLabelText("Password"), "x")
    expect(
      screen.queryByText("Incorrect username or password"),
    ).not.toBeInTheDocument()
  })
})
