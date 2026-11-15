import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import {
  createProductCategory,
  deleteProductCategory,
  listProductCategories,
  updateProductCategory,
} from "@/features/merchant-products/api/categories"
import { listInventory } from "@/features/merchant-products/api/inventory"
import {
  createProduct,
  deleteProduct,
  getProduct,
  listProducts,
  updateProduct,
} from "@/features/merchant-products/api/products"
import {
  createProductSKU,
  deleteProductSKU,
  listProductSKUs,
  updateProductSKU,
  updateProductSKUStock,
} from "@/features/merchant-products/api/skus"
import type {
  Product,
  ProductCategory,
  ProductSKU,
} from "@/features/merchant-products/types"
import { server } from "@/test/server"

type RecordedRequest = {
  method: string
  pathname: string
  search: Record<string, string>
  authorization: string | null
  body: unknown
}

const authToken = ["merchant", "token"].join("-")

const product: Product = {
  id: 7,
  name: "Cotton T-shirt",
  description: "Soft cotton",
  status: "off_shelf",
  category_id: 3,
  category_name: "Clothing",
  sku_count: 0,
  total_stock: 0,
  min_price_cents: null,
  max_price_cents: null,
  created_at: "2026-11-15T08:00:00Z",
  updated_at: "2026-11-15T08:00:00Z",
}

const category: ProductCategory = {
  id: 3,
  name: "Clothing",
  product_count: 0,
  created_at: "2026-11-15T08:00:00Z",
  updated_at: "2026-11-15T08:00:00Z",
}

const sku: ProductSKU = {
  id: 11,
  product_id: 7,
  specs: { Size: "M" },
  price_cents: 1990,
  stock: 5,
  is_enabled: true,
  created_at: "2026-11-15T08:00:00Z",
  updated_at: "2026-11-15T08:00:00Z",
}

describe("merchant product APIs", () => {
  it("uses authenticated REST endpoints and JSON payloads for every operation", async () => {
    const requests: RecordedRequest[] = []
    server.use(
      http.all("*/api/v1/merchant/*", async ({ request }) => {
        const url = new URL(request.url)
        const hasBody =
          request.method !== "GET" && request.method !== "DELETE"
        requests.push({
          method: request.method,
          pathname: url.pathname,
          search: Object.fromEntries(url.searchParams),
          authorization: request.headers.get("Authorization"),
          body: hasBody ? await request.json() : null,
        })

        if (url.pathname.endsWith("/inventory")) {
          return HttpResponse.json([])
        }
        if (url.pathname.endsWith("/categories")) {
          return HttpResponse.json(
            request.method === "GET" ? [category] : category,
          )
        }
        if (url.pathname.endsWith("/skus")) {
          return HttpResponse.json(request.method === "GET" ? [sku] : sku)
        }
        if (url.pathname.endsWith("/products")) {
          return HttpResponse.json(
            request.method === "GET"
              ? {
                  items: [product],
                  page: 2,
                  page_size: 20,
                  total: 1,
                  total_pages: 1,
                }
              : product,
          )
        }
        return HttpResponse.json(
          url.pathname.includes("/skus/") ? sku : product,
        )
      }),
    )

    await listProducts(authToken, {
      page: 2,
      page_size: 20,
      sort: "created_at",
      order: "desc",
      search: "shirt",
      status: "on_shelf",
      category_id: 3,
    })
    await getProduct(authToken, 7)
    await createProduct(authToken, {
      name: "Cotton T-shirt",
      description: "Soft cotton",
      category_id: 3,
      status: "off_shelf",
    })
    await updateProduct(authToken, 7, {
      name: "Cotton T-shirt",
      description: "Soft cotton",
      category_id: null,
      status: "on_shelf",
    })
    await deleteProduct(authToken, 7)

    await listProductSKUs(authToken, 7)
    await createProductSKU(authToken, 7, {
      specs: { Size: "M" },
      price_cents: 1990,
      stock: 5,
      is_enabled: true,
    })
    await updateProductSKU(authToken, 7, 11, {
      specs: { Size: "L" },
      price_cents: 2090,
      stock: 4,
      is_enabled: false,
    })
    await updateProductSKUStock(authToken, 7, 11, 9)
    await deleteProductSKU(authToken, 7, 11)

    await listProductCategories(authToken)
    await createProductCategory(authToken, { name: "Clothing" })
    await updateProductCategory(authToken, 3, { name: "Apparel" })
    await deleteProductCategory(authToken, 3)
    await listInventory(authToken)

    expect(
      requests.map(({ method, pathname }) => ({ method, pathname })),
    ).toEqual([
      { method: "GET", pathname: "/api/v1/merchant/products" },
      { method: "GET", pathname: "/api/v1/merchant/products/7" },
      { method: "POST", pathname: "/api/v1/merchant/products" },
      { method: "PATCH", pathname: "/api/v1/merchant/products/7" },
      { method: "DELETE", pathname: "/api/v1/merchant/products/7" },
      { method: "GET", pathname: "/api/v1/merchant/products/7/skus" },
      { method: "POST", pathname: "/api/v1/merchant/products/7/skus" },
      { method: "PATCH", pathname: "/api/v1/merchant/products/7/skus/11" },
      { method: "PATCH", pathname: "/api/v1/merchant/products/7/skus/11" },
      { method: "DELETE", pathname: "/api/v1/merchant/products/7/skus/11" },
      { method: "GET", pathname: "/api/v1/merchant/categories" },
      { method: "POST", pathname: "/api/v1/merchant/categories" },
      { method: "PATCH", pathname: "/api/v1/merchant/categories/3" },
      { method: "DELETE", pathname: "/api/v1/merchant/categories/3" },
      { method: "GET", pathname: "/api/v1/merchant/inventory" },
    ])
    expect(
      requests.every(
        (request) => request.authorization === `Bearer ${authToken}`,
      ),
    ).toBe(true)
    expect(requests[0]?.search).toEqual({
      page: "2",
      page_size: "20",
      sort: "created_at",
      order: "desc",
      search: "shirt",
      status: "on_shelf",
      category_id: "3",
    })
    expect(requests[2]?.body).toEqual({
      name: "Cotton T-shirt",
      description: "Soft cotton",
      category_id: 3,
      status: "off_shelf",
    })
    expect(requests[3]?.body).toMatchObject({ category_id: null })
    expect(requests[8]?.body).toEqual({ stock: 9 })
  })
})
