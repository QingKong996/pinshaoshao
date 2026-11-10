import { useState } from "react"
import { CheckCircle2, Package, Plus, SearchX } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { ConfirmDialog } from "@/components/merchant/confirm-dialog"
import { EmptyState } from "@/components/merchant/empty-state"
import { PageHeader } from "@/components/merchant/page-header"
import {
  ProductFilters,
  UNCATEGORIZED_FILTER,
  type ProductStatusFilter,
} from "@/components/merchant/products/product-filters"
import { ProductFormDialog } from "@/components/merchant/products/product-form-dialog"
import { ProductTable } from "@/components/merchant/products/product-table"
import { QueryError } from "@/components/merchant/query-error"
import { TableSkeleton } from "@/components/merchant/table-skeleton"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button, buttonVariants } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { Product } from "@/lib/merchant-api"
import { deleteProduct, listCategories, listProducts } from "@/lib/merchant-api"
import { describeMerchantError } from "@/lib/merchant-errors"
import { useDocumentTitle } from "@/lib/use-document-title"
import { useMerchantSession } from "@/lib/use-merchant-session"
import { useRequest } from "@/lib/use-request"
import { cn } from "@/lib/utils"

type ProductsFeedback = "created" | "updated" | "deleted"

export default function ProductsPage() {
  const { t } = useTranslation("merchant")
  useDocumentTitle(t("documentTitle.products"))
  const { token } = useMerchantSession()

  const products = useRequest(() => listProducts(token), [token])
  const categories = useRequest(() => listCategories(token), [token])

  const [search, setSearch] = useState("")
  const [categoryFilter, setCategoryFilter] = useState("")
  const [statusFilter, setStatusFilter] = useState<ProductStatusFilter>("")
  const [formOpen, setFormOpen] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Product | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState<unknown>(null)
  const [feedback, setFeedback] = useState<ProductsFeedback | null>(null)

  const loadError = products.error ?? categories.error
  const loading = products.loading && !products.data
  const query = search.trim().toLowerCase()
  const filteredProducts = (products.data ?? []).filter((product) => {
    if (query && !product.name.toLowerCase().includes(query)) {
      return false
    }
    if (statusFilter && product.status !== statusFilter) {
      return false
    }
    if (categoryFilter === UNCATEGORIZED_FILTER) {
      return product.category_id === null
    }
    if (categoryFilter && product.category_id !== Number(categoryFilter)) {
      return false
    }
    return true
  })

  function reloadAll() {
    products.reload()
    categories.reload()
  }

  function openCreate() {
    setEditingProduct(null)
    setFormOpen(true)
  }

  function openEdit(product: Product) {
    setEditingProduct(product)
    setFormOpen(true)
  }

  function openDelete(product: Product) {
    setDeleteTarget(product)
    setDeleteError(null)
  }

  function handleSaved(kind: "created" | "updated") {
    setFormOpen(false)
    setFeedback(kind)
    products.reload()
  }

  async function handleDelete() {
    if (!deleteTarget) {
      return
    }
    setDeletePending(true)
    setDeleteError(null)
    try {
      await deleteProduct(token, deleteTarget.id)
      setDeleteTarget(null)
      setFeedback("deleted")
      products.reload()
    } catch (error) {
      setDeleteError(error)
    } finally {
      setDeletePending(false)
    }
  }

  function feedbackText(kind: ProductsFeedback): string {
    if (kind === "created") {
      return t("products.created")
    }
    if (kind === "updated") {
      return t("products.updated")
    }
    return t("products.deleted")
  }

  const deleteDescription = deleteTarget
    ? t("products.delete.description", { name: deleteTarget.name })
    : undefined

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("products.title")}
        description={t("products.description")}
        actions={
          <>
            <Button size="sm" onClick={openCreate}>
              <Plus aria-hidden />
              {t("products.create")}
            </Button>
            <Link
              to="/merchant/categories"
              className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
            >
              {t("products.manageCategories")}
            </Link>
          </>
        }
      />

      {loadError ? (
        <QueryError error={loadError} onRetry={reloadAll} />
      ) : null}

      {feedback ? (
        <Alert>
          <CheckCircle2 aria-hidden />
          <AlertDescription>{feedbackText(feedback)}</AlertDescription>
        </Alert>
      ) : null}

      {products.data || !loadError ? (
        <Card>
          <CardContent className="flex flex-col gap-4">
            <ProductFilters
              search={search}
              onSearchChange={setSearch}
              category={categoryFilter}
              onCategoryChange={setCategoryFilter}
              status={statusFilter}
              onStatusChange={setStatusFilter}
              categories={categories.data ?? []}
            />

            {loading ? <TableSkeleton columns={8} /> : null}

            {products.data ? (
              products.data.length === 0 ? (
                <EmptyState
                  icon={Package}
                  title={t("products.empty.title")}
                  description={t("products.empty.description")}
                />
              ) : filteredProducts.length === 0 ? (
                <EmptyState
                  icon={SearchX}
                  title={t("products.emptyFiltered.title")}
                  description={t("products.emptyFiltered.description")}
                />
              ) : (
                <ProductTable
                  products={filteredProducts}
                  onEdit={openEdit}
                  onDelete={openDelete}
                />
              )
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      <ProductFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        token={token}
        categories={categories.data ?? []}
        product={editingProduct}
        onSaved={handleSaved}
      />

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open) {
            if (deletePending) {
              return
            }
            setDeleteTarget(null)
            setDeleteError(null)
          }
        }}
        title={t("products.delete.title")}
        description={
          deleteError ? (
            <span className="flex flex-col gap-1">
              <span>{deleteDescription}</span>
              <span className="font-medium text-destructive">
                {describeMerchantError(deleteError, t)}
              </span>
            </span>
          ) : (
            deleteDescription
          )
        }
        confirmLabel={
          deletePending ? t("common.deleting") : t("common.delete")
        }
        destructive
        pending={deletePending}
        onConfirm={() => void handleDelete()}
      />
    </div>
  )
}
