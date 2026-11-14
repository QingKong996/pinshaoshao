import { useState } from "react"
import { useTranslation } from "react-i18next"
import type { CatalogChanged } from "../types"
import { PRODUCT_NOT_FOUND_ERROR } from "../model/errors"
import { useProduct, useProductCategories } from "../model/queries"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { ProductDetailHeader, ProductDetailBackLink } from "./product-detail-header"
import { ProductDetailSkeleton } from "./product-detail-skeleton"
import { ProductFormDialog } from "./product-form-dialog"
import { ProductMeta } from "./product-meta"
import { QueryError } from "./query-error"
import { SKUSection } from "./sku-section"

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
  const categories = useProductCategories(token, validProductID)

  useDocumentTitle(
    product.data
      ? t("documentTitle.productDetail", { name: product.data.name })
      : t("documentTitle.products"),
  )

  const [productDialogOpen, setProductDialogOpen] = useState(false)

  return (
    <div className="flex flex-col gap-6">
      <ProductDetailBackLink backHref={backHref} />

      {!validProductID ? (
        <QueryError error={PRODUCT_NOT_FOUND_ERROR} />
      ) : null}

      {validProductID && product.error ? (
        <QueryError error={product.error} onRetry={product.refetch} />
      ) : null}

      {validProductID && product.isPending ? <ProductDetailSkeleton /> : null}

      {validProductID && product.data ? (
        <>
          <ProductDetailHeader
            product={product.data}
            onEdit={() => setProductDialogOpen(true)}
          />
          <ProductMeta product={product.data} />
          <SKUSection
            token={token}
            productID={productID}
            onCatalogChanged={onCatalogChanged}
          />
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
    </div>
  )
}
