import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import type { Address, AddressInput } from "@/features/address/types"
import { AddressBookPage } from "@/features/address/ui/address-book-page"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"

const existingAddress: Address = {
  id: 1,
  receiver_name: "Alice",
  receiver_phone: "1111111111",
  shipping_address: "Shanghai",
  is_default: true,
  created_at: "2026-11-27T12:00:00Z",
  updated_at: "2026-11-27T12:00:00Z",
}

describe("AddressBookPage", () => {
  it("lists saved addresses and creates another one", async () => {
    const user = userEvent.setup()
    const addresses = [existingAddress]
    const requestBody = vi.fn<(body: AddressInput) => void>()

    server.use(
      http.get("*/api/v1/addresses", () =>
        HttpResponse.json(addresses),
      ),
      http.post("*/api/v1/addresses", async ({ request }) => {
        const input = (await request.json()) as AddressInput
        requestBody(input)
        const created: Address = {
          id: 2,
          ...input,
          created_at: "2026-11-27T13:00:00Z",
          updated_at: "2026-11-27T13:00:00Z",
        }
        addresses.push(created)
        return HttpResponse.json(created, { status: 201 })
      }),
    )

    renderWithProviders(<AddressBookPage token="token" />)

    expect(await screen.findByText("Alice")).toBeInTheDocument()
    expect(screen.getByText("Default")).toBeInTheDocument()

    await user.click(screen.getByRole("button", { name: "Add address" }))
    await user.type(screen.getByLabelText("Recipient"), "Bob")
    await user.type(screen.getByLabelText("Phone"), "2222222222")
    await user.type(screen.getByLabelText("Full address"), "Beijing")
    await user.click(
      screen.getByLabelText("Make this the default address"),
    )
    await user.click(screen.getByRole("button", { name: "Save address" }))

    await waitFor(() => {
      expect(requestBody).toHaveBeenCalledOnce()
    })
    expect(requestBody.mock.calls[0][0]).toEqual({
      receiver_name: "Bob",
      receiver_phone: "2222222222",
      shipping_address: "Beijing",
      is_default: true,
    })
    expect(await screen.findByText("Bob")).toBeInTheDocument()
  })
})
