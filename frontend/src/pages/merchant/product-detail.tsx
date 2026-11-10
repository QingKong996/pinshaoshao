import { useMemo, useState } from "react"
import { CheckCircle2, PackageSearch, Plus } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link, useParams } from "react-router"
import { ConfirmDialog } from "@/components/merchant/confirm-dialog"
import { EmptyState } from "@/components/merchant/empty-state"
import { PageHeader } from "@/components/merchant/page-header"
import { QueryError } from "@/components/merchant/query-error"
import { SKUFormDialog } from "@/components/merchant/skus/sku-form-dialog"
import { SKUTable } from "@/components/merchant/skus/sku-table"
import { TableSkeleton } from "@/components/merchant/table-skeleton"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button, buttonVariants } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { ApiError } from "@/lib/api"
import { formatDateTime } from "@/lib/format"
import {
  deleteProductSKU,
  getProduct,
  listProductSKUs,
  type SKU,
} from "@/lib/merchant-api"
import { describeMerchantError } from "@/lib/merchant-errors"
import { useDocumentTitle } from "@/lib/use-document-title"
import { useMerchantSession } from "@/lib/use-merchant-session"
import { useRequest } from "@/lib/use-request"
import { cn } from "@/lib/utils"

type DialogState = { mode: "create" } | { mode: "edit"; sku: SKU }

const FEEDBACK_KEYS = {
  created: "skus.created",
  updated: "skus.updated",
  deleted: "skus.deleted",
} as const

type Feedback = keyof typeof FEEDBACK_KEYS

export default function ProductDetailPage() {
  const { t } = useTranslation("merchant")
  const { productID: productIDParam } = useParams()
  const parsedProductID =
    productIDParam !== undefined && /^\d+$/.test(productIDParam)
      ? Number(productIDParam)
      : Number.NaN
  const validProductID =
    Number.isSafeInteger(parsedProductID) && parsedProductID > 0
  const productID = validProductID ? parsedProductID : 0

  const { token } = useMerchantSession()
  const product = useRequest(
    () =>
      validProductID
        ? getProduct(token, productID)
        : Promise.reject(new ApiError(404, "product not found")),
    [token, productID, validProductID],
  )
  const skus = useRequest(
    () =>
      validProductID
        ? listProductSKUs(token, productID)
        : Promise.resolve<SKU[]>([]),
    [token, productID, validProductID],
  )

  const [dialog, setDialog] = useState<DialogState | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<SKU | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState<unknown>(null)
  const [feedback, setFeedback] = useState<Feedback | null>(null)

  const productData = product.data
  const templateKeys = useMemo(() => {
    const keys = new Set<string>()
    for (const sku of skus.data ?? []) {
      for (const key of Object.keys(sku.specs)) {
        keys.add(key)
      }
    }
    return [...keys]
  }, [skus.data])

  useDocumentTitle(
    productData
      ? t("documentTitle.productDetail", { name: productData.name })
      : t("documentTitle.products"),
  )

  const notFound =
    !validProductID ||
    (product.error instanceof ApiError && product.error.status === 404)

  if (notFound) {
    return (
      <div className="flex flex-col gap-6">
        <h1 className="sr-only">{t("productDetail.notFound.title")}</h1>
        <EmptyState
          icon={PackageSearch}
          title={t("productDetail.notFound.title")}
          description={t("productDetail.notFound.description")}
          action={
            <Link
              to="/merchant/products"
              className={cn(
                buttonVariants({ variant: "outline", size: "sm" }),
              )}
            >
              {t("productDetail.back")}
            </Link>
          }
        />
      </div>
    )
  }

  function openCreateDialog() {
    setFeedback(null)
    setDialog({ mode: "create" })
  }

  function openDeleteDialog(sku: SKU) {
    setFeedback(null)
    setDeleteError(null)
    setDeleteTarget(sku)
  }

  async function handleDelete() {
    if (!deleteTarget) {
      return
    }
    setDeleting(true)
    setDeleteError(null)
    try {
      await deleteProductSKU(token, productID, deleteTarget.id)
      setDeleteTarget(null)
      setFeedback("deleted")
      skus.reload()
    } catch (error) {
      setDeleteError(error)
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        backTo="/merchant/products"
        backLabel={t("productDetail.back")}
        title={
          productData
            ? productData.name
            : product.error
              ? t("common.loadError")
              : t("common.loading")
        }
        description={
          productData
            ? productData.description || t("productDetail.info.noDescription")
            : undefined
        }
        actions={
          productData ? (
            <Badge
              variant={
                productData.status === "on_shelf" ? "secondary" : "outline"
              }
            >
              {productData.status === "on_shelf"
                ? t("status.onShelf")
                : t("status.offShelf")}
            </Badge>
          ) : undefined
        }
      />

      {product.error ? (
        <QueryError error={product.error} onRetry={product.reload} />
      ) : null}

      {product.loading && !productData ? (
        <Card>
          <CardContent className="flex flex-col gap-3">
            <Skeleton className="h-4 w-2/3" />
            <Skeleton className="h-4 w-1/2" />
            <Skeleton className="h-4 w-3/5" />
          </CardContent>
        </Card>
      ) : null}

      {productData ? (
        <Card>
          <CardContent className="grid gap-6 md:grid-cols-2">
            <dl className="flex flex-col gap-3 text-sm">
              <div className="flex items-center justify-between gap-3">
                <dt className="text-muted-foreground">
                  {t("productDetail.info.category")}
                </dt>
                <dd>
                  {productData.category_name || t("common.uncategorized")}
                </dd>
              </div>
              <div className="flex items-center justify-between gap-3">
                <dt className="text-muted-foreground">
                  {t("productDetail.info.status")}
                </dt>
                <dd>
                  <Badge
                    variant={
                      productData.status === "on_shelf"
                        ? "secondary"
                        : "outline"
                    }
                  >
                    {productData.status === "on_shelf"
                      ? t("status.onShelf")
                      : t("status.offShelf")}
                  </Badge>
                </dd>
              </div>
              <div className="flex items-center justify-between gap-3">
                <dt className="text-muted-foreground">
                  {t("productDetail.info.createdAt")}
                </dt>
                <dd>{formatDateTime(productData.created_at)}</dd>
              </div>
              <div className="flex items-center justify-between gap-3">
                <dt className="text-muted-foreground">
                  {t("productDetail.info.updatedAt")}
                </dt>
                <dd>{formatDateTime(productData.updated_at)}</dd>
              </div>
            </dl>
            <div className="flex flex-col gap-1.5">
              <span className="text-sm text-muted-foreground">
                {t("productDetail.info.description")}
              </span>
              <p className="text-sm leading-relaxed whitespace-pre-line">
                {productData.description ||
                  t("productDetail.info.noDescription")}
              </p>
            </div>
          </CardContent>
        </Card>
      ) : null}

      {productData ? (
        <Card>
          <CardHeader>
            <CardTitle>{t("skus.title")}</CardTitle>
            <CardDescription>{t("skus.description")}</CardDescription>
            <CardAction>
              <Button
                size="sm"
                className="gap-1.5"
                onClick={openCreateDialog}
                disabled={!skus.data}
              >
                <Plus aria-hidden />
                {t("skus.add")}
              </Button>
            </CardAction>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {feedback ? (
              <Alert>
                <CheckCircle2 aria-hidden />
                <AlertDescription>{t(FEEDBACK_KEYS[feedback])}</AlertDescription>
              </Alert>
            ) : null}

            {skus.error ? (
              <QueryError error={skus.error} onRetry={skus.reload} />
            ) : null}

            {skus.loading && !skus.data ? (
              <TableSkeleton columns={6} />
            ) : null}

            {skus.data && skus.data.length === 0 ? (
              <EmptyState
                icon={PackageSearch}
                title={t("skus.empty.title")}
                description={t("skus.empty.description")}
                action={
                  <Button size="sm" onClick={openCreateDialog}>
                    {t("skus.add")}
                  </Button>
                }
                className="py-10"
              />
            ) : null}

            {skus.data && skus.data.length > 0 ? (
              <SKUTable
                skus={skus.data}
                onEdit={(sku) => {
                  setFeedback(null)
                  setDialog({ mode: "edit", sku })
                }}
                onDelete={openDeleteDialog}
              />
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      {dialog ? (
        <SKUFormDialog
          open
          onOpenChange={(open) => {
            if (!open) {
              setDialog(null)
            }
          }}
          productID={productID}
          sku={dialog.mode === "edit" ? dialog.sku : null}
          templateKeys={templateKeys}
          onSaved={() => {
            setFeedback(dialog.mode === "create" ? "created" : "updated")
            setDialog(null)
            skus.reload()
          }}
        />
      ) : null}

      {deleteTarget ? (
        <ConfirmDialog
          open
          onOpenChange={(open) => {
            if (!open) {
              if (deleting) {
                return
              }
              setDeleteTarget(null)
              setDeleteError(null)
            }
          }}
          title={t("skus.delete.title")}
          description={
            deleteError ? (
              <span className="text-destructive">
                {describeMerchantError(deleteError, t)}
              </span>
            ) : (
              t("skus.delete.description")
            )
          }
          destructive
          pending={deleting}
          onConfirm={() => void handleDelete()}
        />
      ) : null}
    </div>
  )
}
