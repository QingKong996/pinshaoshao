import { useDeferredValue, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { Card } from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import type { CatalogChanged, InventoryItem } from "../types"
import { useInventory } from "../model/queries"
import {
  ALL,
  DEFAULT_PAGE_SIZE,
  filterInventory,
  type StockFilter,
} from "../model/inventory"
import { InventoryFilters } from "./inventory-filters"
import { InventoryPagination } from "./inventory-pagination"
import { InventoryTable } from "./inventory-table"
import { QueryError } from "./query-error"
import { StockFormDialog } from "./stock-form-dialog"

const EMPTY_INVENTORY: InventoryItem[] = []

export default function MerchantConsoleInventoryPage({
  token,
  productHref,
  onCatalogChanged,
}: {
  token: string
  productHref: (productID: number) => string
  onCatalogChanged: CatalogChanged
}) {
  const { t } = useTranslation("console")
  useDocumentTitle(t("documentTitle.inventory"))

  const [search, setSearch] = useState("")
  const [categoryFilter, setCategoryFilter] = useState(ALL)
  const [stockFilter, setStockFilter] = useState<StockFilter>(ALL)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState<number>(DEFAULT_PAGE_SIZE)
  const [stockTarget, setStockTarget] = useState<InventoryItem | null>(null)

  const deferredSearch = useDeferredValue(search)
  const inventoryQuery = useInventory(token)
  const items = inventoryQuery.data ?? EMPTY_INVENTORY
  const filteredItems = useMemo(
    () =>
      filterInventory(items, {
        search: deferredSearch,
        categoryFilter,
        stockFilter,
      }),
    [categoryFilter, deferredSearch, items, stockFilter],
  )
  const categoryItems = useMemo(() => {
    const names = new Map<number, string>()
    for (const item of items) {
      if (item.category_id !== null && item.category_name !== null) {
        names.set(item.category_id, item.category_name)
      }
    }
    return [...names.entries()]
      .sort(([, left], [, right]) => left.localeCompare(right))
      .map(([id, name]) => ({ value: String(id), label: name }))
  }, [items])

  const totalPages = Math.ceil(filteredItems.length / pageSize)
  const visiblePage = Math.min(page, Math.max(1, totalPages))
  const pageItems = filteredItems.slice(
    (visiblePage - 1) * pageSize,
    visiblePage * pageSize,
  )
  const hasFilters =
    deferredSearch.trim() !== "" ||
    categoryFilter !== ALL ||
    stockFilter !== ALL

  function changeSearch(nextSearch: string) {
    setSearch(nextSearch)
    setPage(1)
  }

  function changeCategoryFilter(nextCategory: string) {
    setCategoryFilter(nextCategory)
    setPage(1)
  }

  function changeStockFilter(nextStock: StockFilter) {
    setStockFilter(nextStock)
    setPage(1)
  }

  function changePageSize(nextPageSize: number) {
    setPageSize(nextPageSize)
    setPage(1)
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("inventory.title")} />

      {inventoryQuery.error && !inventoryQuery.data ? (
        <QueryError
          error={inventoryQuery.error}
          onRetry={inventoryQuery.refetch}
        />
      ) : (
        <Card className="gap-0 overflow-hidden p-0">
          <InventoryFilters
            search={search}
            onSearchChange={changeSearch}
            categoryItems={categoryItems}
            categoryFilter={categoryFilter}
            onCategoryFilterChange={changeCategoryFilter}
            stockFilter={stockFilter}
            onStockFilterChange={changeStockFilter}
          />

          {inventoryQuery.error ? (
            <div className="px-4 pt-4">
              <QueryError
                error={inventoryQuery.error}
                onRetry={inventoryQuery.refetch}
              />
            </div>
          ) : null}

          <InventoryTable
            items={pageItems}
            isPending={inventoryQuery.isPending}
            isEmpty={!hasFilters && items.length === 0}
            productHref={productHref}
            onAdjustStock={setStockTarget}
          />

          <InventoryPagination
            page={visiblePage}
            pageSize={pageSize}
            total={filteredItems.length}
            totalPages={totalPages}
            onPageChange={setPage}
            onPageSizeChange={changePageSize}
          />
        </Card>
      )}

      <StockFormDialog
        token={token}
        onCatalogChanged={onCatalogChanged}
        item={stockTarget}
        onClose={() => setStockTarget(null)}
      />
    </div>
  )
}
