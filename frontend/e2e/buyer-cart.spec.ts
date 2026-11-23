import { expect, test, type Locator, type Page } from "@playwright/test"

const cartItems = [
  {
    id: 1,
    product_id: 10,
    sku_id: 5,
    product_name: "Cotton T-shirt",
    sku_specs: { Color: "Blue", Size: "M" },
    unit_price_cents: 1999,
    quantity: 2,
    line_total_cents: 3998,
    availability: "available",
    created_at: "2026-11-19T08:00:00Z",
    updated_at: "2026-11-19T08:00:00Z",
  },
] as const

const summaryInset = 20

async function mockBuyerApi(page: Page) {
  await page.route("**/api/v1/**", async (route) => {
    const request = route.request()
    const pathname = new URL(request.url()).pathname

    if (request.method() === "POST" && pathname.endsWith("/login")) {
      await route.fulfill({ json: { token: "buyer-session" } })
      return
    }
    if (request.method() === "GET" && pathname.endsWith("/feed")) {
      await route.fulfill({
        json: {
          items: [],
          page: 1,
          page_size: 20,
          total: 0,
          total_pages: 0,
        },
      })
      return
    }
    if (request.method() === "GET" && pathname.endsWith("/cart")) {
      await route.fulfill({ json: cartItems })
      return
    }

    await route.fulfill({
      status: 404,
      json: { error: `unhandled ${request.method()} ${pathname}` },
    })
  })
}

async function signInAndOpenCart(page: Page) {
  await page.goto("/signin")
  await page.getByLabel("Username").fill("buyer")
  await page.getByLabel("Password", { exact: true }).fill("password")
  await page.getByRole("button", { name: "Sign in", exact: true }).click()
  await page.waitForURL(/\/$/)

  await page.goto("/cart")
  await expect(
    page.getByRole("heading", { name: "Shopping cart", exact: true }),
  ).toBeVisible()
  await expect(page.getByText("Cotton T-shirt")).toBeVisible()
}

function summaryRegion(page: Page) {
  return page.getByRole("region", { name: "Selected items" })
}

async function boundingBox(locator: Locator) {
  const box = await locator.boundingBox()
  if (!box) {
    throw new Error("Expected element to have a bounding box")
  }
  return box
}

async function expectInsideCard(
  card: Locator,
  content: Locator,
  inset: number,
) {
  const cardBox = await boundingBox(card)
  const contentBox = await boundingBox(content)

  expect(contentBox.x).toBeGreaterThanOrEqual(cardBox.x + inset)
  expect(contentBox.x + contentBox.width).toBeLessThanOrEqual(
    cardBox.x + cardBox.width - inset,
  )
  expect(contentBox.y).toBeGreaterThanOrEqual(cardBox.y)
  expect(contentBox.y + contentBox.height).toBeLessThanOrEqual(
    cardBox.y + cardBox.height,
  )
}

async function expectVerticalGap(
  upper: Locator,
  lower: Locator,
  minGap: number,
) {
  const upperBox = await boundingBox(upper)
  const lowerBox = await boundingBox(lower)

  expect(lowerBox.y - (upperBox.y + upperBox.height)).toBeGreaterThanOrEqual(
    minGap,
  )
}

async function expectNoHorizontalOverflow(page: Page) {
  const dimensions = await page.evaluate(() => ({
    clientWidth: document.documentElement.clientWidth,
    scrollWidth: document.documentElement.scrollWidth,
  }))

  expect(dimensions.scrollWidth).toBeLessThanOrEqual(dimensions.clientWidth)
}

test("keeps the cart summary readable across desktop widths", async ({
  page,
}) => {
  await mockBuyerApi(page)
  await signInAndOpenCart(page)

  const summary = summaryRegion(page)
  const item = page
    .getByRole("listitem")
    .filter({ hasText: "Cotton T-shirt" })
  const title = summary.getByRole("heading", {
    name: "Selected items",
  })
  const totals = summary.locator("dl")
  const notice = summary.getByText(/Prices shown are locked/)
  const checkout = summary.getByRole("button", { name: "Checkout" })

  await expect(summary).toBeVisible()

  for (const width of [1024, 1440]) {
    await page.setViewportSize({ width, height: 900 })
    await expectNoHorizontalOverflow(page)

    const summaryBox = await boundingBox(summary)
    const itemBox = await boundingBox(item)
    expect(summaryBox.width).toBeGreaterThanOrEqual(340)
    expect(
      summaryBox.x - (itemBox.x + itemBox.width),
    ).toBeGreaterThanOrEqual(24)

    await expectInsideCard(summary, title, summaryInset)
    await expectInsideCard(summary, totals, summaryInset)
    await expectInsideCard(summary, notice, summaryInset)
    await expectInsideCard(summary, checkout, summaryInset)
    await expectVerticalGap(title, totals, 16)
    await expectVerticalGap(totals, notice, 16)
    await expectVerticalGap(notice, checkout, 16)

    const checkoutBox = await boundingBox(checkout)
    expect(checkoutBox.height).toBeGreaterThanOrEqual(32)
  }
})

test("stacks the cart summary cleanly on mobile", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await mockBuyerApi(page)
  await signInAndOpenCart(page)

  const summary = summaryRegion(page)
  const item = page
    .getByRole("listitem")
    .filter({ hasText: "Cotton T-shirt" })
  const title = summary.getByRole("heading", {
    name: "Selected items",
  })
  const checkout = summary.getByRole("button", { name: "Checkout" })

  await expect(summary).toBeVisible()
  await expectNoHorizontalOverflow(page)

  const summaryBox = await boundingBox(summary)
  const itemBox = await boundingBox(item)
  expect(summaryBox.y).toBeGreaterThanOrEqual(
    itemBox.y + itemBox.height + 16,
  )
  expect(summaryBox.x).toBeGreaterThanOrEqual(16)
  expect(summaryBox.x + summaryBox.width).toBeLessThanOrEqual(374)

  await expectInsideCard(summary, title, summaryInset)
  await expectInsideCard(summary, checkout, summaryInset)
  await expectVerticalGap(title, summary.locator("dl"), 16)
  await expect(checkout).toBeInViewport()
})
