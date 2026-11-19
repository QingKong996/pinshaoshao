import { describe, expect, it } from "vitest"
import {
  ALL,
  UNCATEGORIZED,
  buildProductListQuery,
  defaultOrderForSort,
} from "@/features/merchant-products/model/filters"

describe("defaultOrderForSort", () => {
  it("uses ascending order for text columns and descending for dates", () => {
    expect(defaultOrderForSort("name")).toBe("asc")
    expect(defaultOrderForSort("category")).toBe("asc")
    expect(defaultOrderForSort("status")).toBe("asc")
    expect(defaultOrderForSort("created_at")).toBe("desc")
  })
})

describe("buildProductListQuery", () => {
  const input = {
    page: 2,
    pageSize: 20,
    sort: "name" as const,
    order: "asc" as const,
  }
  const expectedBase = {
    page: 2,
    page_size: 20,
    sort: "name" as const,
    order: "asc" as const,
  }

  it("omits inactive filters", () => {
    expect(
      buildProductListQuery({
        ...input,
        search: "   ",
        statusFilter: ALL,
        categoryFilter: ALL,
      }),
    ).toEqual(expectedBase)
  })

  it("adds normalized search, status, and category filters", () => {
    expect(
      buildProductListQuery({
        ...input,
        search: "  shirt ",
        statusFilter: "on_shelf",
        categoryFilter: "3",
      }),
    ).toEqual({
      ...expectedBase,
      search: "shirt",
      status: "on_shelf",
      category_id: 3,
    })
  })

  it("requests uncategorized products and ignores invalid category values", () => {
    expect(
      buildProductListQuery({
        ...input,
        search: "",
        statusFilter: ALL,
        categoryFilter: UNCATEGORIZED,
      }).uncategorized,
    ).toBe(true)
    expect(
      buildProductListQuery({
        ...input,
        search: "",
        statusFilter: ALL,
        categoryFilter: "not-a-number",
      }),
    ).toEqual(expectedBase)
  })
})
