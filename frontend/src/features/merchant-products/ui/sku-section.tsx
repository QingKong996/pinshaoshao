import { useCallback, useState } from "react"
import { AlertCircle, Plus } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import type { ProductSKU, CatalogChanged } from "../types"
import { useDeleteProductSKU } from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import { useProductSKUs } from "../model/queries"
import { DeleteSKUDialog } from "./delete-sku-dialog"
import { QueryError } from "./query-error"
import { SKUFormDialog } from "./sku-form-dialog"
import { SKUTable } from "./sku-table"

const EMPTY_SKUS: ProductSKU[] = []

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

  function openCreateSKU() {
    setEditingSKU(null)
    setSkuDialogOpen(true)
  }

  const openEditSKU = useCallback((sku: ProductSKU) => {
    setEditingSKU(sku)
    setSkuDialogOpen(true)
  }, [])

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
          {deleteProductSKUMutation.error ? (
            <Alert variant="destructive">
              <AlertCircle aria-hidden />
              <AlertTitle>{t("common.actionError")}</AlertTitle>
              <AlertDescription>
                {describeConsoleError(deleteProductSKUMutation.error, t)}
              </AlertDescription>
            </Alert>
          ) : null}

          {skus.error ? (
            <QueryError error={skus.error} onRetry={skus.refetch} />
          ) : null}

          <SKUTable
            skus={skus.data ?? EMPTY_SKUS}
            isPending={skus.isPending}
            showEmpty={Boolean(skus.data)}
            onEdit={openEditSKU}
            onDelete={setDeletingSKU}
          />
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
