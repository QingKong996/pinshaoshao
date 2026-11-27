import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import { createMockAuth } from "@/mock/auth/provider"
import { server } from "@/test/server"

const apiBaseURL = "/mock-auth-storage-test"
const storageKey = "pinshaoshao:mock-auth-storage-test"
const mockAuth = createMockAuth({
  apiBaseURL,
  signInPath: "/signin",
  storageKey,
})

function MockAuthHarness() {
  const { session, signIn, signOut } = mockAuth.useMockAuth()

  return (
    <>
      <output aria-label="mock session">
        {session?.username ?? "anonymous"}
      </output>
      <button
        type="button"
        onClick={() => void signIn("operator", "password")}
      >
        Sign in
      </button>
      <button type="button" onClick={() => void signOut()}>
        Sign out
      </button>
    </>
  )
}

describe("createMockAuth", () => {
  it("keeps mock sessions in memory when storage writes fail", async () => {
    const user = userEvent.setup()
    const setItem = vi
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw new DOMException("storage disabled", "SecurityError")
      })
    const removeItem = vi
      .spyOn(Storage.prototype, "removeItem")
      .mockImplementation(() => {
        throw new DOMException("storage disabled", "SecurityError")
      })
    let logoutAuthorization: string | null = null

    server.use(
      http.post("*/mock-auth-storage-test/auth/login", () =>
        HttpResponse.json({ token: "in-memory-mock-token" }),
      ),
      http.post("*/mock-auth-storage-test/auth/logout", ({ request }) => {
        logoutAuthorization = request.headers.get("Authorization")
        return HttpResponse.json({})
      }),
    )

    render(
      <mockAuth.Provider>
        <MockAuthHarness />
      </mockAuth.Provider>,
    )

    await user.click(screen.getByRole("button", { name: "Sign in" }))

    expect(await screen.findByLabelText("mock session")).toHaveTextContent(
      "operator",
    )
    expect(setItem).toHaveBeenCalled()

    await user.click(screen.getByRole("button", { name: "Sign out" }))

    expect(screen.getByLabelText("mock session")).toHaveTextContent(
      "anonymous",
    )
    expect(removeItem).toHaveBeenCalled()
    await waitFor(() => {
      expect(logoutAuthorization).toBe("Bearer in-memory-mock-token")
    })
  })
})
