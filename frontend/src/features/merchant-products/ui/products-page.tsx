import { useDeferredValue, useState } from "react"
import { AlertCircle, Plus } from "lucide-react"
import { useTranslation } from "react-i18next"
import { PageHeader } from "@/shared/ui/page-header"
import { ProductFormDialog } from "./product-form-dialog"
import { ProductDeleteDialog } from "./product-delete-dialog"
import { ProductShelfStatusDialog } from "./product-shelf-status-dialog"
import { ProductsFilters } from "./products-filters"
import { ProductsPagination } from "./products-pagination"
import { ProductsTable } from "./products-table"
import { QueryError } from "./query-error"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import type {
  CatalogChanged,
  Product,
  ProductSort,
  SortOrder,
} from "../types"
import {
  useDeleteProduct,
  useUpdateProduct,
} from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import { useProducts, useProductCategories } from "../model/queries"
import {
  ALL,
  buildProductListQuery,
  DEFAULT_ORDER,
  DEFAULT_PAGE_SIZE,
  DEFAULT_SORT,
  defaultOrderForSort,
  type StatusFilter,
} from "../model/filters"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

type ProductFormState = {
  open: boolean
  product: Product | null
}

export default function MerchantConsoleProductsPage({
  token,
  productHref,
  onCatalogChanged,
}: {
  token: string
  productHref: (productID: number) => string
  onCatalogChanged: CatalogChanged
}) {
  const { t } = useTranslation("console")
  useDocumentTitle(t("documentTitle.products"))

  const [search, setSearch] = useState("")
  const [statusFilter, setStatusFilter] = useState<StatusFilter>(ALL)
  const [categoryFilter, setCategoryFilter] = useState(ALL)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState<number>(DEFAULT_PAGE_SIZE)
  const [sort, setSort] = useState<ProductSort>(DEFAULT_SORT)
  const [order, setOrder] = useState<SortOrder>(DEFAULT_ORDER)
  const [form, setForm] = useState<ProductFormState>({
    open: false,
    product: null,
  })
  const [shelfStatusTarget, setShelfStatusTarget] =
    useState<Product | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Product | null>(null)

  const deferredSearch = useDeferredValue(search)
  const products = useProducts(
    token,
    buildProductListQuery({
      page,
      pageSize,
      sort,
      order,
      search: deferredSearch,
      statusFilter,
      categoryFilter,
    }),
  )
  const categories = useProductCategories(token)
  const updateProductMutation = useUpdateProduct(token, onCatalogChanged)
  const deleteProductMutation = useDeleteProduct(token, onCatalogChanged)

  const actionError =
    updateProductMutation.error ?? deleteProductMutation.error

  const items = products.data?.items ?? []
  const total = products.data?.total ?? 0
  const totalPages = products.data?.total_pages ?? 0
  const hasFilters =
    deferredSearch.trim() !== "" ||
    statusFilter !== ALL ||
    categoryFilter !== ALL

  function changeSearch(nextSearch: string) {
    setSearch(nextSearch)
    setPage(1)
  }

  function changeStatusFilter(nextStatus: StatusFilter) {
    setStatusFilter(nextStatus)
    setPage(1)
  }

  function changeCategoryFilter(nextCategory: string) {
    setCategoryFilter(nextCategory)
    setPage(1)
  }

  function changeSort(nextSort: ProductSort) {
    if (nextSort === sort) {
      setOrder((current) => (current === "asc" ? "desc" : "asc"))
    } else {
      setSort(nextSort)
      setOrder(defaultOrderForSort(nextSort))
    }
    setPage(1)
  }

  function changePageSize(nextPageSize: number) {
    setPageSize(nextPageSize)
    setPage(1)
  }

  function openCreateDialog() {
    setForm({ open: true, product: null })
  }

  function openEditDialog(product: Product) {
    setForm({ open: true, product })
  }

  function confirmShelfStatusChange() {
    if (!shelfStatusTarget) {
      return
    }

    const willBeOnShelf = shelfStatusTarget.status === "off_shelf"
    updateProductMutation.mutate(
      {
        productID: shelfStatusTarget.id,
        input: {
          name: shelfStatusTarget.name,
          description: shelfStatusTarget.description,
          category_id: shelfStatusTarget.category_id,
          status: willBeOnShelf ? "on_shelf" : "off_shelf",
        },
      },
      {
        onSuccess: () => {
          if (statusFilter !== ALL && items.length === 1 && page > 1) {
            setPage(page - 1)
          }
        },
        onSettled: () => setShelfStatusTarget(null),
      },
    )
  }

  function toggleShelfStatus(product: Product) {
    setShelfStatusTarget(product)
  }

  function confirmDelete() {
    if (!deleteTarget) {
      return
    }
    deleteProductMutation.mutate(deleteTarget.id, {
      onSuccess: () => {
        if (items.length === 1 && page > 1) {
          setPage(page - 1)
        }
      },
      onSettled: () => setDeleteTarget(null),
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("products.title")}
        actions={
          <Button onClick={openCreateDialog}>
            <Plus aria-hidden />
            {t("products.new")}
          </Button>
        }
      />

      {products.error && !products.data ? (
        <QueryError error={products.error} onRetry={products.refetch} />
      ) : (
        <Card className="gap-0 overflow-hidden p-0">
          <ProductsFilters
            search={search}
            onSearchChange={changeSearch}
            statusFilter={statusFilter}
            onStatusFilterChange={changeStatusFilter}
            categoryFilter={categoryFilter}
            onCategoryFilterChange={changeCategoryFilter}
            categories={categories.data ?? []}
          />

          {actionError ? (
            <div className="px-4 pt-4">
              <Alert variant="destructive">
                <AlertCircle aria-hidden />
                <AlertTitle>{t("common.actionError")}</AlertTitle>
                <AlertDescription>
                  {describeConsoleError(actionError, t)}
                </AlertDescription>
              </Alert>
            </div>
          ) : null}

          <ProductsTable
            products={items}
            isPending={products.isPending}
            isCatalogEmpty={!hasFilters && total === 0}
            sort={sort}
            order={order}
            onSort={changeSort}
            productHref={productHref}
            onEdit={openEditDialog}
            onToggleShelfStatus={toggleShelfStatus}
            onDelete={setDeleteTarget}
          />

          <ProductsPagination
            page={page}
            pageSize={pageSize}
            total={total}
            totalPages={totalPages}
            onPageChange={setPage}
            onPageSizeChange={changePageSize}
          />
        </Card>
      )}

      <ProductFormDialog
        token={token}
        onCatalogChanged={onCatalogChanged}
        open={form.open}
        onOpenChange={(open) =>
          setForm((current) => ({ ...current, open }))
        }
        product={form.product ?? undefined}
        categories={categories.data ?? []}
      />

      <ProductDeleteDialog
        product={deleteTarget}
        isPending={deleteProductMutation.isPending}
        onClose={() => setDeleteTarget(null)}
        onConfirm={confirmDelete}
      />

      <ProductShelfStatusDialog
        product={shelfStatusTarget}
        isPending={updateProductMutation.isPending}
        onClose={() => setShelfStatusTarget(null)}
        onConfirm={confirmShelfStatusChange}
      />
    </div>
  )
}
