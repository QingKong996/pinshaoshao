import type { Product, ProductStatus } from "../types"

export const ALL = "all"
export const UNCATEGORIZED = "none"
export type StatusFilter = typeof ALL | ProductStatus

export function filterProducts(
  products: Product[],
  search: string,
  statusFilter: StatusFilter,
  categoryFilter: string,
) {
  const query = search.trim().toLowerCase()
  return products.filter((product) => {
    if (statusFilter !== ALL && product.status !== statusFilter) {
      return false
    }
    if (categoryFilter === UNCATEGORIZED) {
      if (product.category_id !== null) {
        return false
      }
    } else if (
      categoryFilter !== ALL &&
      String(product.category_id) !== categoryFilter
    ) {
      return false
    }
    return (
      !query ||
      product.name.toLowerCase().includes(query) ||
      product.description.toLowerCase().includes(query)
    )
  })
}
