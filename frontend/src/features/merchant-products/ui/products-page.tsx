import { useMemo, useState } from "react"
import { AlertCircle, Plus } from "lucide-react"
import { useTranslation } from "react-i18next"
import { PageHeader } from "@/shared/ui/page-header"
import { ProductFormDialog } from "./product-form-dialog"
import { ProductDeleteDialog } from "./product-delete-dialog"
import { ProductsFilters } from "./products-filters"
import { ProductsTable } from "./products-table"
import { QueryError } from "./query-error"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import type { Product, CatalogChanged } from "../types"
import {
  useDeleteProduct,
  useUpdateProduct,
} from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import { useProducts, useProductCategories } from "../model/queries"
import {
  ALL,
  filterProducts,
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

  const products = useProducts(token)
  const categories = useProductCategories(token)
  const updateProductMutation = useUpdateProduct(token, onCatalogChanged)
  const deleteProductMutation = useDeleteProduct(token, onCatalogChanged)

  const [search, setSearch] = useState("")
  const [statusFilter, setStatusFilter] = useState<StatusFilter>(ALL)
  const [categoryFilter, setCategoryFilter] = useState(ALL)
  const [form, setForm] = useState<ProductFormState>({
    open: false,
    product: null,
  })
  const [deleteTarget, setDeleteTarget] = useState<Product | null>(null)

  const actionError =
    updateProductMutation.error ?? deleteProductMutation.error

  const filteredProducts = useMemo(
    () => filterProducts(products.data ?? [], search, statusFilter, categoryFilter),
    [products.data, search, statusFilter, categoryFilter],
  )

  function openCreateDialog() {
    setForm({ open: true, product: null })
  }

  function openEditDialog(product: Product) {
    setForm({ open: true, product })
  }

  function toggleShelfStatus(product: Product) {
    updateProductMutation.mutate({
      productID: product.id,
      input: {
        name: product.name,
        description: product.description,
        category_id: product.category_id,
        status: product.status === "on_shelf" ? "off_shelf" : "on_shelf",
      },
    })
  }

  function confirmDelete() {
    if (!deleteTarget) {
      return
    }
    deleteProductMutation.mutate(deleteTarget.id, {
      onSettled: () => setDeleteTarget(null),
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("products.title")}
        description={t("products.description")}
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
            onSearchChange={setSearch}
            statusFilter={statusFilter}
            onStatusFilterChange={setStatusFilter}
            categoryFilter={categoryFilter}
            onCategoryFilterChange={setCategoryFilter}
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
            products={filteredProducts}
            isPending={products.isPending}
            isCatalogEmpty={products.data?.length === 0}
            productHref={productHref}
            onEdit={openEditDialog}
            onToggleShelfStatus={toggleShelfStatus}
            onDelete={setDeleteTarget}
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
    </div>
  )
}
