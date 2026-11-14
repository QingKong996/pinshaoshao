import { useDeferredValue, useMemo, useState } from "react"
import { AlertCircle, Plus, Search } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import { Input } from "@/shared/ui/input"
import { PageHeader } from "@/shared/ui/page-header"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import type { CatalogChanged, ProductCategory } from "../types"
import { useDeleteProductCategory } from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import { useProductCategories } from "../model/queries"
import { CategoriesTable } from "./categories-table"
import { CategoryDeleteDialog } from "./category-delete-dialog"
import { CategoryFormDialog } from "./category-form-dialog"
import { QueryError } from "./query-error"

const EMPTY_CATEGORIES: ProductCategory[] = []

export default function MerchantConsoleCategoriesPage({
  token,
  onCatalogChanged,
}: {
  token: string
  onCatalogChanged: CatalogChanged
}) {
  const { t } = useTranslation("console")
  useDocumentTitle(t("documentTitle.categories"))

  const [search, setSearch] = useState("")
  const [form, setForm] = useState<{
    open: boolean
    category: ProductCategory | null
  }>({ open: false, category: null })
  const [deleteTarget, setDeleteTarget] = useState<ProductCategory | null>(
    null,
  )

  const deferredSearch = useDeferredValue(search)
  const categoriesQuery = useProductCategories(token)
  const deleteCategoryMutation = useDeleteProductCategory(
    token,
    onCatalogChanged,
  )
  const categories = categoriesQuery.data ?? EMPTY_CATEGORIES
  const normalizedSearch = deferredSearch.trim().toLocaleLowerCase()
  const filteredCategories = useMemo(
    () =>
      normalizedSearch
        ? categories.filter((category) =>
            category.name.toLocaleLowerCase().includes(normalizedSearch),
          )
        : categories,
    [categories, normalizedSearch],
  )

  function openCreateDialog() {
    setForm({ open: true, category: null })
  }

  function openEditDialog(category: ProductCategory) {
    setForm({ open: true, category })
  }

  function confirmDelete() {
    if (!deleteTarget) {
      return
    }
    deleteCategoryMutation.mutate(deleteTarget.id, {
      onSettled: () => setDeleteTarget(null),
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("categories.title")}
        actions={
          <Button onClick={openCreateDialog}>
            <Plus aria-hidden />
            {t("categories.new")}
          </Button>
        }
      />

      {categoriesQuery.error && !categoriesQuery.data ? (
        <QueryError
          error={categoriesQuery.error}
          onRetry={categoriesQuery.refetch}
        />
      ) : (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="flex flex-col gap-2 border-b p-4 sm:flex-row sm:items-center">
            <div className="relative w-full sm:max-w-xs">
              <Search
                className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground"
                aria-hidden
              />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder={t("categories.searchPlaceholder")}
                aria-label={t("categories.searchPlaceholder")}
                className="pl-8"
              />
            </div>
            <p className="text-sm text-muted-foreground sm:ml-auto">
              {t("categories.count", { count: filteredCategories.length })}
            </p>
          </div>

          {categoriesQuery.error ? (
            <div className="px-4 pt-4">
              <QueryError
                error={categoriesQuery.error}
                onRetry={categoriesQuery.refetch}
              />
            </div>
          ) : null}

          {deleteCategoryMutation.error ? (
            <div className="px-4 pt-4">
              <Alert variant="destructive">
                <AlertCircle aria-hidden />
                <AlertTitle>{t("common.actionError")}</AlertTitle>
                <AlertDescription>
                  {describeConsoleError(deleteCategoryMutation.error, t)}
                </AlertDescription>
              </Alert>
            </div>
          ) : null}

          <CategoriesTable
            categories={filteredCategories}
            isPending={categoriesQuery.isPending}
            isCatalogEmpty={
              normalizedSearch === "" && categories.length === 0
            }
            onEdit={openEditDialog}
            onDelete={setDeleteTarget}
          />
        </Card>
      )}

      <CategoryFormDialog
        token={token}
        onCatalogChanged={onCatalogChanged}
        open={form.open}
        onOpenChange={(open) =>
          setForm((current) => ({ ...current, open }))
        }
        category={form.category ?? undefined}
      />

      <CategoryDeleteDialog
        category={deleteTarget}
        isPending={deleteCategoryMutation.isPending}
        onClose={() => setDeleteTarget(null)}
        onConfirm={confirmDelete}
      />
    </div>
  )
}
