import { expect, test, type Page } from "@playwright/test"

const apiPrefix = "/api/v1"
const sessionToken = ["e2e", "session"].join("-")

async function mockMerchantApi(page: Page) {
  await page.route("**/api/v1/**", async (route) => {
    const request = route.request()
    const pathname = new URL(request.url()).pathname

    if (request.method() === "POST" && pathname.endsWith("/merchant/login")) {
      await route.fulfill({ json: { token: sessionToken } })
      return
    }
    if (
      request.method() === "GET" &&
      pathname.endsWith("/merchant/dashboard")
    ) {
      await route.fulfill({
        json: {
          product_count: 1,
          on_shelf_product_count: 0,
          sku_count: 2,
          category_count: 1,
          total_stock: 15,
          inventory_value_cents: 39900,
          out_of_stock_sku_count: 0,
        },
      })
      return
    }
    if (request.method() === "GET" && pathname.endsWith("/merchant/shop")) {
      await route.fulfill({
        json: {
          id: 1,
          username: "merchant",
          name: "Example Shop",
          description: "A test shop",
          email: "shop@example.com",
          status: "active",
          created_at: "2026-11-15T08:00:00Z",
          updated_at: "2026-11-15T09:00:00Z",
        },
      })
      return
    }
    if (request.method() === "GET" && pathname.endsWith("/merchant/orders")) {
      await route.fulfill({
        json: {
          items: [
            {
              id: 4,
              order_id: 12,
              status: "awaiting_shipment",
              total_amount_cents: 3998,
              item_count: 2,
              receiver_name: "Receiver",
              receiver_phone: "1234567890",
              shipping_address: "Shanghai",
              paid_at: "2026-11-20T12:05:00Z",
              created_at: "2026-11-20T12:00:00Z",
            },
          ],
          page: 1,
          page_size: 10,
          total: 1,
          total_pages: 1,
        },
      })
      return
    }
    if (request.method() === "GET" && pathname.endsWith("/merchant/products")) {
      await route.fulfill({
        json: {
          items: [
            {
              id: 7,
              name: "Cotton T-shirt",
              description: "Soft cotton",
              status: "off_shelf",
              category_id: 3,
              category_name: "Clothing",
              sku_count: 2,
              min_price_cents: 1990,
              max_price_cents: 2590,
              created_at: "2026-11-15T08:00:00Z",
              updated_at: "2026-11-15T09:00:00Z",
            },
          ],
          page: 1,
          page_size: 10,
          total: 1,
          total_pages: 1,
        },
      })
      return
    }
    if (
      request.method() === "GET" &&
      pathname.endsWith("/merchant/products/7")
    ) {
      await route.fulfill({
        json: {
          id: 7,
          name: "Cotton T-shirt",
          description: "Soft cotton",
          status: "off_shelf",
          category_id: 3,
          category_name: "Clothing",
          sku_count: 2,
          min_price_cents: 1990,
          max_price_cents: 2590,
          created_at: "2026-11-15T08:00:00Z",
          updated_at: "2026-11-15T09:00:00Z",
        },
      })
      return
    }
    if (
      request.method() === "GET" &&
      pathname.endsWith("/merchant/products/7/skus")
    ) {
      await route.fulfill({ json: [] })
      return
    }
    if (
      request.method() === "POST" &&
      pathname.endsWith("/merchant/products/7/skus")
    ) {
      await route.fulfill({
        json: {
          id: 11,
          product_id: 7,
          specs: { Size: "M" },
          price_cents: 1990,
          is_enabled: true,
          created_at: "2026-11-15T08:00:00Z",
          updated_at: "2026-11-15T08:00:00Z",
        },
      })
      return
    }
    if (request.method() === "GET" && pathname.endsWith("/merchant/categories")) {
      await route.fulfill({
        json: [
          {
            id: 3,
            name: "Clothing",
            product_count: 1,
            created_at: "2026-11-15T08:00:00Z",
            updated_at: "2026-11-15T09:00:00Z",
          },
        ],
      })
      return
    }
    if (
      request.method() === "DELETE" &&
      pathname.endsWith("/merchant/products/7")
    ) {
      await route.fulfill({ status: 204 })
      return
    }

    await route.fulfill({
      status: 404,
      json: { error: `unhandled ${request.method()} ${pathname}` },
    })
  })
}

test("redirects unauthenticated merchants to sign in", async ({ page }) => {
  await page.goto("/merchant/dashboard")

  await expect(page).toHaveURL(/\/merchant\/signin$/)
  await expect(
    page.getByRole("heading", { name: "Merchant sign in" }),
  ).toBeVisible()
})

test("public landing page exposes merchant entry points", async ({ page }) => {
  await page.goto("/merchant")

  await expect(page).toHaveTitle(/拼少少/)
  await expect(page.getByText("Sell on Pinshaoshao")).toBeVisible()
  await expect(
    page.getByRole("link", { name: "Open a shop free" }).first(),
  ).toBeVisible()
})

test("merchant can sign in, inspect products, and confirm deletion", async ({
  page,
}) => {
  await mockMerchantApi(page)
  await page.goto("/merchant/signin")

  await page.getByLabel("Shop username").fill("merchant")
  await page
    .getByRole("textbox", { name: "Password", exact: true })
    .fill("a".repeat(12))
  await page
    .getByRole("button", { name: "Sign in to your shop" })
    .click()

  await expect(page).toHaveURL(/\/merchant\/dashboard$/)
  await expect(
    page.getByRole("heading", { name: "Overview" }),
  ).toBeVisible()
  await expect(page.getByText("Example Shop")).toBeVisible()

  await page.getByRole("link", { name: "Orders" }).click()
  await expect(page).toHaveURL(/\/merchant\/orders$/)
  await expect(
    page.getByRole("heading", { name: "Orders" }),
  ).toBeVisible()
  await expect(page.getByText("Sub-order #4")).toBeVisible()
  await expect(page.getByText("Awaiting shipment")).toBeVisible()

  await page.getByRole("link", { name: "Product info" }).click()
  await expect(page).toHaveURL(/\/merchant\/products$/)
  await expect(
    page.getByRole("link", { name: "Cotton T-shirt" }),
  ).toBeVisible()

  await page.getByRole("link", { name: "Cotton T-shirt" }).click()
  await expect(page).toHaveURL(/\/merchant\/products\/7$/)
  await expect(
    page.getByRole("heading", { name: /Cotton T-shirt/ }),
  ).toBeVisible()

  await page.getByRole("button", { name: "Add SKU" }).click()
  const skuDialog = page.getByRole("dialog")
  await skuDialog.getByLabel("Spec name").fill("Size")
  await skuDialog.getByLabel("Spec value").fill("M")
  await skuDialog.getByLabel("Price (CNY)").fill("19.90")
  const createSkuRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith(`${apiPrefix}/merchant/products/7/skus`),
  )
  await skuDialog.getByRole("button", { name: "Add", exact: true }).click()
  await createSkuRequest

  await page.getByRole("link", { name: "Back to products" }).click()
  await expect(page).toHaveURL(/\/merchant\/products$/)
  await page.getByRole("button", { name: "Delete" }).click()
  const dialog = page.getByRole("alertdialog")
  await expect(dialog).toBeVisible()

  const deleteRequest = page.waitForRequest(
    (request) =>
      request.method() === "DELETE" &&
      request.url().endsWith(`${apiPrefix}/merchant/products/7`),
  )
  await dialog.getByRole("button", { name: "Delete" }).click()
  await deleteRequest
})
