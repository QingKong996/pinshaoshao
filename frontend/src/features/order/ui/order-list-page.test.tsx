import { act, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { orderQueryKeys } from "@/features/order/model/query-keys"
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

  it("returns to the last valid page when the current page is empty", async () => {
    const user = userEvent.setup()
    let pageOneRequests = 0
    const requestedPages: string[] = []
    server.use(
      http.get("*/api/v1/orders", ({ request }) => {
        const page = new URL(request.url).searchParams.get("page") ?? ""
        requestedPages.push(page)
        if (page === "2") {
          return HttpResponse.json({
            items: [],
            page: 2,
            page_size: 10,
            total: 1,
            total_pages: 1,
          })
        }
        pageOneRequests++
        return HttpResponse.json({
          ...orderList,
          page: 1,
          total: pageOneRequests === 1 ? 11 : 1,
          total_pages: pageOneRequests === 1 ? 2 : 1,
        })
      }),
    )

    renderWithProviders(<OrderListPage token="buyer-token" />)

    expect(await screen.findByText("Order #12")).toBeVisible()
    await user.click(screen.getByRole("button", { name: "Next page" }))

    await waitFor(() => {
      expect(requestedPages).toEqual(["1", "2", "1"])
    })
    expect(await screen.findByText("Order #12")).toBeVisible()
    expect(screen.queryByText("No orders yet")).not.toBeInTheDocument()
  })

  it("shows a refresh error without hiding loaded orders", async () => {
    let shouldFail = false
    server.use(
      http.get("*/api/v1/orders", () =>
        shouldFail
          ? HttpResponse.json(
              { error: "temporarily unavailable" },
              { status: 500 },
            )
          : HttpResponse.json(orderList),
      ),
    )
    const { queryClient } = renderWithProviders(
      <OrderListPage token="buyer-token" />,
    )

    expect(await screen.findByText("Order #12")).toBeVisible()
    shouldFail = true

    await act(async () => {
      await queryClient.invalidateQueries({
        queryKey: orderQueryKeys.orderList("buyer-token", 1, 10),
      })
    })

    expect(
      await screen.findByText("Could not load the order"),
    ).toBeVisible()
    expect(screen.getByText("Order #12")).toBeVisible()
  })
})
