import type {
  ProductListQuery,
  ProductSort,
  ProductStatus,
  SortOrder,
} from "../types"

export const ALL = "all"
export const UNCATEGORIZED = "none"
export type StatusFilter = typeof ALL | ProductStatus

export const DEFAULT_PAGE_SIZE = 10
export const PAGE_SIZE_OPTIONS = [10, 20, 50] as const
export const DEFAULT_SORT: ProductSort = "created_at"
export const DEFAULT_ORDER: SortOrder = "desc"

export function defaultOrderForSort(sort: ProductSort): SortOrder {
  switch (sort) {
    case "name":
    case "category":
    case "status":
      return "asc"
    default:
      return "desc"
  }
}

export function buildProductListQuery({
  page,
  pageSize,
  sort,
  order,
  search,
  statusFilter,
  categoryFilter,
}: {
  page: number
  pageSize: number
  sort: ProductSort
  order: SortOrder
  search: string
  statusFilter: StatusFilter
  categoryFilter: string
}) {
  const query: ProductListQuery = {
    page,
    page_size: pageSize,
    sort,
    order,
  }
  const trimmedSearch = search.trim()
  if (trimmedSearch) {
    query.search = trimmedSearch
  }
  if (statusFilter !== ALL) {
    query.status = statusFilter
  }
  if (categoryFilter === UNCATEGORIZED) {
    query.uncategorized = true
  } else if (categoryFilter !== ALL) {
    const categoryID = Number(categoryFilter)
    if (Number.isInteger(categoryID) && categoryID > 0) {
      query.category_id = categoryID
    }
  }
  return query
}
