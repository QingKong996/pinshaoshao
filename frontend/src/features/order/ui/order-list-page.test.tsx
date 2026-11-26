import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { orderList } from "@/test/commerce-fixtures"
import { OrderListPage } from "./order-list-page"

describe("OrderListPage", () => {
  it("renders orders and loads the next page", async () => {
    const user = userEvent.setup()
    let requestedPage = ""
    server.use(
      http.get("*/api/v1/orders", ({ request }) => {
        requestedPage = new URL(request.url).searchParams.get("page") ?? ""
        return HttpResponse.json({
          ...orderList,
          page: Number(requestedPage),
          total: 11,
          total_pages: 2,
        })
      }),
    )

    renderWithProviders(<OrderListPage token="buyer-token" />)

    expect(await screen.findByText("Order #12")).toBeVisible()
    expect(screen.getByText("CN¥39.98")).toBeVisible()
    expect(screen.getByText("2 item(s) · 1 shop(s)")).toBeVisible()

    await user.click(screen.getByRole("button", { name: "Next page" }))
    await waitFor(() => {
      expect(requestedPage).toBe("2")
    })
    expect(
      await screen.findByText("Page 2 of 2"),
    ).toBeVisible()
  })

  it("renders an empty state", async () => {
    server.use(
      http.get("*/api/v1/orders", () =>
        HttpResponse.json({
          items: [],
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        }),
      ),
    )

    renderWithProviders(<OrderListPage token="buyer-token" />)

    expect(await screen.findByText("No orders yet")).toBeVisible()
    expect(
      screen.getByRole("link", { name: "Start shopping" }),
    ).toHaveAttribute("href", "/")
  })
})
