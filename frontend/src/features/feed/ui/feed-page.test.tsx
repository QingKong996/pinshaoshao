import { screen, waitFor } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { FeedPage } from "./feed-page"

const useInViewMock = vi.hoisted(() => vi.fn())

vi.mock("react-intersection-observer", () => ({
  useInView: useInViewMock,
}))

function feedItem(id: number, name: string) {
  return {
    id,
    shop_id: 1,
    shop_name: "Example shop",
    name,
    description: "In stock",
    category_name: "Food",
    min_price_cents: 1999,
    max_price_cents: 2999,
    available_stock: 8,
  }
}

describe("FeedPage", () => {
  beforeEach(() => {
    useInViewMock.mockReturnValue({
      ref: () => undefined,
      inView: true,
    })
  })

  it("loads the next page automatically before the feed bottom", async () => {
    const requests: number[] = []
    const authorizations: Array<string | null> = []
    server.use(
      http.get("*/api/v1/feed", ({ request }) => {
        const url = new URL(request.url)
        const page = Number(url.searchParams.get("page"))
        requests.push(page)
        authorizations.push(request.headers.get("Authorization"))

        if (page === 1) {
          return HttpResponse.json({
            items: [feedItem(1, "Product 1"), feedItem(2, "Product 2")],
            page: 1,
            page_size: 20,
            total: 3,
            total_pages: 2,
          })
        }
        return HttpResponse.json({
          items: [feedItem(3, "Product 3")],
          page: 2,
          page_size: 20,
          total: 3,
          total_pages: 2,
        })
      }),
    )

    renderWithProviders(<FeedPage token="token" />)

    expect(await screen.findByText("Product 1")).toBeInTheDocument()
    expect(await screen.findByText("Product 3")).toBeInTheDocument()
    await waitFor(() => expect(requests).toEqual([1, 2]))
    expect(authorizations).toEqual(["Bearer token", "Bearer token"])
    expect(screen.getByText("Showing 3 of 3 products")).toBeInTheDocument()
    expect(
      screen.getAllByRole("img", { name: "Product image placeholder" }),
    ).toHaveLength(3)
    expect(
      screen.queryByRole("button", { name: "Load more" }),
    ).not.toBeInTheDocument()
    expect(useInViewMock).toHaveBeenCalledWith({
      rootMargin: "0px 0px 1200px 0px",
    })
  })

  it("shows the load error when the first page fails", async () => {
    server.use(
      http.get(
        "*/api/v1/feed",
        () => new HttpResponse(null, { status: 500 }),
      ),
    )

    renderWithProviders(<FeedPage token="token" />)

    expect(await screen.findByText("Could not load your feed")).toBeInTheDocument()
    expect(
      await screen.findByText(
        "The feed is temporarily unavailable. Please try again.",
      ),
    ).toBeInTheDocument()
  })
})
