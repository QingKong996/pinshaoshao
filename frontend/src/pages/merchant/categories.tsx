import { useState } from "react"
import { CheckCircle2, Plus, Tags } from "lucide-react"
import { useTranslation } from "react-i18next"
import { CategoryFormDialog } from "@/components/merchant/categories/category-form-dialog"
import { CategoryTable } from "@/components/merchant/categories/category-table"
import { ConfirmDialog } from "@/components/merchant/confirm-dialog"
import { EmptyState } from "@/components/merchant/empty-state"
import { PageHeader } from "@/components/merchant/page-header"
import { QueryError } from "@/components/merchant/query-error"
import { TableSkeleton } from "@/components/merchant/table-skeleton"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { Category } from "@/lib/merchant-api"
import { deleteCategory, listCategories } from "@/lib/merchant-api"
import { describeMerchantError } from "@/lib/merchant-errors"
import { useDocumentTitle } from "@/lib/use-document-title"
import { useMerchantSession } from "@/lib/use-merchant-session"
import { useRequest } from "@/lib/use-request"

type CategoriesFeedback = "created" | "updated" | "deleted"

export default function CategoriesPage() {
  const { t } = useTranslation("merchant")
  useDocumentTitle(t("documentTitle.categories"))
  const { token } = useMerchantSession()

  const categories = useRequest(() => listCategories(token), [token])

  const [formOpen, setFormOpen] = useState(false)
  const [editingCategory, setEditingCategory] = useState<Category | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Category | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState<unknown>(null)
  const [feedback, setFeedback] = useState<CategoriesFeedback | null>(null)

  const loading = categories.loading && !categories.data

  function openCreate() {
    setEditingCategory(null)
    setFormOpen(true)
  }

  function openRename(category: Category) {
    setEditingCategory(category)
    setFormOpen(true)
  }

  function openDelete(category: Category) {
    setDeleteTarget(category)
    setDeleteError(null)
  }

  function handleSaved(kind: "created" | "updated") {
    setFormOpen(false)
    setFeedback(kind)
    categories.reload()
  }

  async function handleDelete() {
    if (!deleteTarget) {
      return
    }
    setDeletePending(true)
    setDeleteError(null)
    try {
      await deleteCategory(token, deleteTarget.id)
      setDeleteTarget(null)
      setFeedback("deleted")
      categories.reload()
    } catch (error) {
      setDeleteError(error)
    } finally {
      setDeletePending(false)
    }
  }

  function feedbackText(kind: CategoriesFeedback): string {
    if (kind === "created") {
      return t("categories.created")
    }
    if (kind === "updated") {
      return t("categories.updated")
    }
    return t("categories.deleted")
  }

  const deleteDescription = deleteTarget
    ? t("categories.delete.description", { name: deleteTarget.name })
    : undefined

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("categories.title")}
        description={t("categories.description")}
        actions={
          <Button size="sm" onClick={openCreate}>
            <Plus aria-hidden />
            {t("categories.create")}
          </Button>
        }
      />

      {categories.error ? (
        <QueryError error={categories.error} onRetry={categories.reload} />
      ) : null}

      {feedback ? (
        <Alert>
          <CheckCircle2 aria-hidden />
          <AlertDescription>{feedbackText(feedback)}</AlertDescription>
        </Alert>
      ) : null}

      {categories.data || !categories.error ? (
        <Card>
          <CardContent>
            {loading ? <TableSkeleton columns={4} /> : null}

            {categories.data ? (
              categories.data.length === 0 ? (
                <EmptyState
                  icon={Tags}
                  title={t("categories.empty.title")}
                  description={t("categories.empty.description")}
                />
              ) : (
                <CategoryTable
                  categories={categories.data}
                  onRename={openRename}
                  onDelete={openDelete}
                />
              )
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      <CategoryFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        token={token}
        category={editingCategory}
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
        title={t("categories.delete.title")}
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
