import { useState } from "react"
import { AlertCircle, Pencil, Plus, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
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
import { formatPriceCents } from "@/shared/lib/format"
import type { ProductSKU, CatalogChanged } from "../types"
import { useDeleteProductSKU } from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import { useProductSKUs } from "../model/queries"
import { DeleteSKUDialog } from "./delete-sku-dialog"
import { formatSpecsLabel } from "./format-specs"
import { QueryError } from "./query-error"
import { SKUFormDialog } from "./sku-form-dialog"

export function SKUSection({
  token,
  productID,
  onCatalogChanged,
}: {
  token: string
  productID: number
  onCatalogChanged: CatalogChanged
}) {
  const { t } = useTranslation("console")

  const skus = useProductSKUs(token, productID)
  const deleteProductSKUMutation = useDeleteProductSKU(token, onCatalogChanged)

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
    <>
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
                <TableHead>{t("productDetail.skus.table.specs")}</TableHead>
                <TableHead>{t("productDetail.skus.table.price")}</TableHead>
                <TableHead>{t("productDetail.skus.table.stock")}</TableHead>
                <TableHead>{t("productDetail.skus.table.enabled")}</TableHead>
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
                    <Badge variant={sku.is_enabled ? "secondary" : "outline"}>
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

      {skuDialogOpen ? (
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
    </>
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
