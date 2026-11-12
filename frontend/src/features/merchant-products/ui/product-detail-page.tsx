import { useState } from "react"
import {
  AlertCircle,
  ArrowLeft,
  LoaderCircle,
  Pencil,
  Plus,
  Trash2,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { PageHeader } from "@/shared/ui/page-header"
import { ProductFormDialog } from "./product-form-dialog"
import { ProductStatusBadge } from "./product-status-badge"
import { QueryError } from "./query-error"
import { SKUFormDialog } from "./sku-form-dialog"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog"
import { Badge } from "@/shared/ui/badge"
import { Button } from "@/shared/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { Skeleton } from "@/shared/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import {
  formatDateTime,
  formatPriceCents,
  formatPriceRange,
} from "@/shared/lib/format"
import type { Product, ProductSKU, CatalogChanged } from "../types"
import { useDeleteProductSKU } from "../model/mutations"
import { describeConsoleError, PRODUCT_NOT_FOUND_ERROR } from "../model/errors"
import { useProduct, useProductSKUs, useProductCategories } from "../model/queries"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

export default function MerchantConsoleProductDetailPage({
  token,
  productID,
  backHref,
  onCatalogChanged,
}: {
  token: string
  productID: number
  backHref: string
  onCatalogChanged: CatalogChanged
}) {
  const { t } = useTranslation("console")
  const validProductID = Number.isInteger(productID) && productID > 0

  const product = useProduct(token, productID)
  const skus = useProductSKUs(token, productID)
  const categories = useProductCategories(token, validProductID)
  const deleteProductSKUMutation = useDeleteProductSKU(token, onCatalogChanged)

  useDocumentTitle(
    product.data
      ? t("documentTitle.productDetail", { name: product.data.name })
      : t("documentTitle.products"),
  )

  const [productDialogOpen, setProductDialogOpen] = useState(false)
  const [skuDialogOpen, setSkuDialogOpen] = useState(false)
  const [editingSKU, setEditingSKU] = useState<ProductSKU | null>(null)
  const [deletingSKU, setDeletingSKU] = useState<ProductSKU | null>(null)

  const actionError = deleteProductSKUMutation.error

  function openCreateSKU() {
    setEditingSKU(null)
    setSkuDialogOpen(true)
  }

  function openEditSKU(sku: ProductSKU) {
    setEditingSKU(sku)
    setSkuDialogOpen(true)
  }

  function closeSKUDialog() {
    setSkuDialogOpen(false)
    setEditingSKU(null)
  }

  function handleDeleteSKU() {
    if (!deletingSKU) {
      return
    }
    deleteProductSKUMutation.mutate(
      { productID, skuID: deletingSKU.id },
      { onSettled: () => setDeletingSKU(null) },
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <Link
        to={backHref}
        className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden />
        {t("productDetail.back")}
      </Link>

      {!validProductID ? (
        <QueryError error={PRODUCT_NOT_FOUND_ERROR} />
      ) : null}

      {validProductID && product.error ? (
        <QueryError error={product.error} onRetry={product.refetch} />
      ) : null}

      {validProductID && product.isPending ? <ProductDetailSkeleton /> : null}

      {validProductID && product.data ? (
        <>
          <PageHeader
            title={
              <span className="flex flex-wrap items-center gap-2">
                {product.data.name}
                <ProductStatusBadge status={product.data.status} />
              </span>
            }
            description={
              product.data.description || t("productDetail.noDescription")
            }
            actions={
              <Button onClick={() => setProductDialogOpen(true)}>
                <Pencil aria-hidden />
                {t("productDetail.edit")}
              </Button>
            }
          />

          <ProductMeta product={product.data} />

          <Card>
            <CardHeader>
              <CardTitle>{t("productDetail.skus.title")}</CardTitle>
              <CardDescription>
                {t("productDetail.skus.description")}
              </CardDescription>
              <CardAction>
                <Button
                  size="sm"
                  onClick={openCreateSKU}
                  disabled={!skus.data}
                >
                  <Plus aria-hidden />
                  {t("productDetail.skus.add")}
                </Button>
              </CardAction>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              {actionError ? (
                <Alert variant="destructive">
                  <AlertCircle aria-hidden />
                  <AlertTitle>{t("common.actionError")}</AlertTitle>
                  <AlertDescription>
                    {describeConsoleError(actionError, t)}
                  </AlertDescription>
                </Alert>
              ) : null}

              {skus.error ? (
                <QueryError error={skus.error} onRetry={skus.refetch} />
              ) : null}

              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>
                      {t("productDetail.skus.table.specs")}
                    </TableHead>
                    <TableHead>
                      {t("productDetail.skus.table.price")}
                    </TableHead>
                    <TableHead>
                      {t("productDetail.skus.table.stock")}
                    </TableHead>
                    <TableHead>
                      {t("productDetail.skus.table.enabled")}
                    </TableHead>
                    <TableHead className="text-right">
                      {t("productDetail.skus.table.actions")}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {skus.isPending ? <SKUTableSkeleton /> : null}

                  {skus.data && skus.data.length === 0 ? (
                    <TableRow className="hover:bg-transparent">
                      <TableCell colSpan={5} className="py-8">
                        <div className="flex flex-col items-center gap-1 text-center">
                          <span className="font-medium">
                            {t("productDetail.skus.empty.title")}
                          </span>
                          <span className="text-sm text-muted-foreground">
                            {t("productDetail.skus.empty.description")}
                          </span>
                        </div>
                      </TableCell>
                    </TableRow>
                  ) : null}

                  {skus.data?.map((sku) => (
                    <TableRow key={sku.id}>
                      <TableCell className="whitespace-normal">
                        {formatSpecsLabel(
                          sku,
                          t("productDetail.skus.defaultSpecs"),
                        )}
                      </TableCell>
                      <TableCell className="tabular-nums">
                        {formatPriceCents(sku.price_cents)}
                      </TableCell>
                      <TableCell className="tabular-nums">{sku.stock}</TableCell>
                      <TableCell>
                        <Badge
                          variant={sku.is_enabled ? "secondary" : "outline"}
                        >
                          {sku.is_enabled
                            ? t("productDetail.skus.enabled")
                            : t("productDetail.skus.disabled")}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={t("productDetail.skus.actions.edit")}
                            onClick={() => openEditSKU(sku)}
                          >
                            <Pencil aria-hidden />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={t("productDetail.skus.actions.delete")}
                            onClick={() => setDeletingSKU(sku)}
                          >
                            <Trash2 aria-hidden />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </>
      ) : null}

      <ProductFormDialog
        token={token}
        onCatalogChanged={onCatalogChanged}
        open={productDialogOpen}
        onOpenChange={setProductDialogOpen}
        product={product.data ?? undefined}
        categories={categories.data ?? []}
      />

      {skuDialogOpen && validProductID ? (
        <SKUFormDialog
          token={token}
          onCatalogChanged={onCatalogChanged}
          key={editingSKU ? `edit-${editingSKU.id}` : "create"}
          open
          onOpenChange={(open) => {
            if (!open) {
              closeSKUDialog()
            }
          }}
          productID={productID}
          sku={editingSKU ?? undefined}
          skus={skus.data ?? []}
        />
      ) : null}

      <DeleteSKUDialog
        sku={deletingSKU}
        pending={deleteProductSKUMutation.isPending}
        onOpenChange={(open) => {
          if (!open && !deleteProductSKUMutation.isPending) {
            setDeletingSKU(null)
          }
        }}
        onConfirm={handleDeleteSKU}
      />
    </div>
  )
}

function ProductMeta({ product }: { product: Product }) {
  const { t } = useTranslation("console")
  const priceRange = formatPriceRange(
    product.min_price_cents,
    product.max_price_cents,
  )
  const items = [
    {
      label: t("productDetail.meta.category"),
      value: product.category_name ?? t("products.filters.uncategorized"),
    },
    {
      label: t("productDetail.meta.skuCount"),
      value: String(product.sku_count),
    },
    {
      label: t("productDetail.meta.stock"),
      value: String(product.total_stock),
    },
    { label: t("productDetail.meta.price"), value: priceRange ?? "—" },
    {
      label: t("productDetail.meta.updatedAt"),
      value: formatDateTime(product.updated_at),
    },
  ]

  return (
    <Card>
      <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        {items.map((item) => (
          <div key={item.label} className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{item.label}</span>
            <span className="font-medium tabular-nums">{item.value}</span>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}

function DeleteSKUDialog({
  sku,
  pending,
  onOpenChange,
  onConfirm,
}: {
  sku: ProductSKU | null
  pending: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  const { t } = useTranslation("console")
  const specs = sku
    ? formatSpecsLabel(sku, t("productDetail.skus.defaultSpecs"))
    : ""

  return (
    <AlertDialog open={sku !== null} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {t("productDetail.skus.deleteDialog.title")}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {t("productDetail.skus.deleteDialog.description", { specs })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={pending}>
            {t("common.cancel")}
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={pending}
            onClick={onConfirm}
          >
            {pending ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : null}
            {t("productDetail.skus.deleteDialog.confirm")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

function ProductDetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-72" />
      </div>
      <Card>
        <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
          {Array.from({ length: 5 }, (_, index) => (
            <div key={index} className="flex flex-col gap-1.5">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-5 w-24" />
            </div>
          ))}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-24" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          {Array.from({ length: 3 }, (_, index) => (
            <Skeleton key={index} className="h-9 w-full" />
          ))}
        </CardContent>
      </Card>
    </div>
  )
}

function SKUTableSkeleton() {
  return (
    <>
      {Array.from({ length: 3 }, (_, index) => (
        <TableRow key={index}>
          <TableCell>
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-16" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-10" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-5 w-16" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-6 w-16" />
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}

function formatSpecsLabel(sku: ProductSKU, emptyLabel: string) {
  const entries = Object.entries(sku.specs).sort(([left], [right]) =>
    left.localeCompare(right),
  )
  if (entries.length === 0) {
    return emptyLabel
  }
  return entries.map(([key, value]) => `${key}: ${value}`).join(" / ")
}
