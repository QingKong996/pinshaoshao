import type { InventoryItem } from "../types"

export const ALL = "all"
export type StockFilter = typeof ALL | "in_stock" | "out_of_stock"

export function filterInventory(
  items: InventoryItem[],
  {
    search,
    categoryFilter,
    stockFilter,
  }: {
    search: string
    categoryFilter: string
    stockFilter: StockFilter
  },
) {
  const normalizedSearch = search.trim().toLocaleLowerCase()

  return items.filter((item) => {
    if (
      categoryFilter !== ALL &&
      String(item.category_id) !== categoryFilter
    ) {
      return false
    }
    if (stockFilter === "in_stock" && item.available_stock === 0) {
      return false
    }
    if (stockFilter === "out_of_stock" && item.available_stock > 0) {
      return false
    }
    if (!normalizedSearch) {
      return true
    }

    return [
      item.product_name,
      item.category_name ?? "",
      ...Object.entries(item.specs).flatMap(([key, value]) => [key, value]),
    ].some((value) => value.toLocaleLowerCase().includes(normalizedSearch))
  })
}
