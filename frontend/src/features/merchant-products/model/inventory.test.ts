import { describe, expect, it } from "vitest"
import {
  ALL,
  filterInventory,
} from "@/features/merchant-products/model/inventory"
import type { InventoryItem } from "@/features/merchant-products/types"

const items: InventoryItem[] = [
  {
    sku_id: 1,
    product_id: 10,
    product_name: "Cotton T-shirt",
    category_id: 2,
    category_name: "Clothing",
    specs: { Color: "Blue", Size: "M" },
    price_cents: 1990,
    stock: 8,
    is_enabled: true,
    updated_at: "2026-11-15T08:00:00Z",
  },
  {
    sku_id: 2,
    product_id: 11,
    product_name: "Canvas Bag",
    category_id: null,
    category_name: null,
    specs: {},
    price_cents: 990,
    stock: 0,
    is_enabled: false,
    updated_at: "2026-11-15T09:00:00Z",
  },
]

describe("filterInventory", () => {
  it("filters by category and stock state", () => {
    expect(
      filterInventory(items, {
        search: "",
        categoryFilter: "2",
        stockFilter: "in_stock",
      }).map((item) => item.sku_id),
    ).toEqual([1])
    expect(
      filterInventory(items, {
        search: "",
        categoryFilter: ALL,
        stockFilter: "out_of_stock",
      }).map((item) => item.sku_id),
    ).toEqual([2])
  })

  it("searches product, category, and spec values case-insensitively", () => {
    expect(
      filterInventory(items, {
        search: "  blue ",
        categoryFilter: ALL,
        stockFilter: ALL,
      }).map((item) => item.sku_id),
    ).toEqual([1])
    expect(
      filterInventory(items, {
        search: "clothing",
        categoryFilter: ALL,
        stockFilter: ALL,
      }).map((item) => item.sku_id),
    ).toEqual([1])
    expect(
      filterInventory(items, {
        search: "bag",
        categoryFilter: ALL,
        stockFilter: ALL,
      }).map((item) => item.sku_id),
    ).toEqual([2])
  })
})
