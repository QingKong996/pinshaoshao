import { useState } from "react"
import {
  ArrowLeft,
  CheckCircle2,
  LoaderCircle,
  PackageOpen,
  ShoppingCart,
  Sparkles,
  Zap,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Alert, AlertDescription } from "@/shared/ui/alert"
import { Badge } from "@/shared/ui/badge"
import { Button, buttonVariants } from "@/shared/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { Input } from "@/shared/ui/input"
import { QueryError } from "@/shared/ui/query-error"
import { Separator } from "@/shared/ui/separator"
import { Skeleton } from "@/shared/ui/skeleton"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { formatPriceCents } from "@/shared/lib/format"
import {
  describeProductActionError,
  describeProductError,
} from "../model/errors"
import { isSKUPurchasable } from "../model/product"
import { useProductDetail } from "../model/queries"
import type { ProductDetail, ProductSKU } from "../types"
import { formatProductSpecs } from "./format-specs"
import { ProductSKUPicker } from "./product-sku-picker"

export function ProductDetailPage({
  token,
  productID,
  onAddToCart,
}: {
  token: string
  productID: number
  onAddToCart: (skuID: number, quantity: number) => Promise<unknown>
}) {
  const { t } = useTranslation("feed")
  const validProductID = Number.isInteger(productID) && productID > 0
  const productQuery = useProductDetail(token, productID, validProductID)
  const [selectedSKUID, setSelectedSKUID] = useState<number | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [added, setAdded] = useState(false)
  const [isAdding, setIsAdding] = useState(false)
  const [actionError, setActionError] = useState<unknown>(null)

  useDocumentTitle(
    productQuery.data
      ? t("detail.documentTitle", { name: productQuery.data.name })
      : t("title"),
  )

  if (!validProductID) {
    return (
      <QueryError
        title={t("errors.productLoadTitle")}
        description={t("errors.invalidProduct")}
        action={
          <Link
            to="/"
            className={buttonVariants({ variant: "outline", size: "sm" })}
          >
            <ArrowLeft aria-hidden />
            {t("detail.back")}
          </Link>
        }
      />
    )
  }

  if (productQuery.error) {
    return (
      <QueryError
        title={t("errors.productLoadTitle")}
        description={describeProductError(productQuery.error, t)}
        retryLabel={t("errors.retry")}
        onRetry={productQuery.refetch}
      />
    )
  }

  if (productQuery.isPending || !productQuery.data) {
    return <ProductDetailSkeleton />
  }

  return (
    <ProductDetailContent
      product={productQuery.data}
      selectedSKUID={selectedSKUID}
      quantity={quantity}
      added={added}
      actionError={actionError}
      isAdding={isAdding}
      onSelect={(sku) => {
        setSelectedSKUID(sku.id)
        setQuantity(1)
        setAdded(false)
        setActionError(null)
      }}
      onQuantityChange={(next) => {
        setQuantity(next)
        setAdded(false)
        setActionError(null)
      }}
      onAddToCart={async (sku) => {
        setAdded(false)
        setActionError(null)
        setIsAdding(true)
        try {
          await onAddToCart(sku.id, quantity)
          setAdded(true)
        } catch (error) {
          setActionError(error)
        } finally {
          setIsAdding(false)
        }
      }}
    />
  )
}

function ProductDetailContent({
  product,
  selectedSKUID,
  quantity,
  added,
  actionError,
  isAdding,
  onSelect,
  onQuantityChange,
  onAddToCart,
}: {
  product: ProductDetail
  selectedSKUID: number | null
  quantity: number
  added: boolean
  actionError: unknown
  isAdding: boolean
  onSelect: (sku: ProductSKU) => void
  onQuantityChange: (quantity: number) => void
  onAddToCart: (sku: ProductSKU) => Promise<void>
}) {
  const { t } = useTranslation("feed")
  const firstPurchasable = product.skus.find(isSKUPurchasable)
  const selectedSKU =
    product.skus.find((sku) => sku.id === selectedSKUID) ?? firstPurchasable
  const purchasable = selectedSKU ? isSKUPurchasable(selectedSKU) : false
  const maxQuantity =
    selectedSKU && selectedSKU.available_stock > 0
      ? selectedSKU.available_stock
      : 1
  const directCheckoutHref =
    selectedSKU && purchasable
      ? `/checkout?product_id=${product.id}&sku_id=${selectedSKU.id}&quantity=${quantity}`
      : null

  return (
    <div className="flex flex-col gap-6">
      <Link
        to="/"
        className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden />
        {t("detail.back")}
      </Link>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
        <Card className="overflow-hidden p-0">
          <div
            role="img"
            aria-label={t("imagePlaceholder")}
            className="flex aspect-[4/5] items-center justify-center bg-muted"
          >
            <PackageOpen
              className="size-16 text-muted-foreground/50"
              aria-hidden
            />
          </div>
        </Card>

        <div className="flex min-w-0 flex-col gap-5">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="secondary">
              {product.category_name ?? t("uncategorized")}
            </Badge>
            <span className="text-sm text-muted-foreground">
              {product.shop_name}
            </span>
          </div>

          <div className="space-y-2">
            <h1 className="font-heading text-3xl font-semibold tracking-tight text-balance">
              {product.name}
            </h1>
            {product.description ? (
              <p className="max-w-2xl leading-relaxed text-muted-foreground">
                {product.description}
              </p>
            ) : null}
          </div>

          <div className="flex items-baseline gap-2">
            <span className="font-heading text-2xl font-semibold tabular-nums">
              {selectedSKU
                ? formatPriceCents(selectedSKU.price_cents)
                : t("detail.skus.unavailable")}
            </span>
            {selectedSKU ? (
              <span className="text-sm text-muted-foreground">
                {formatProductSpecs(
                  selectedSKU.specs,
                  t("detail.skus.noSpecs"),
                )}
              </span>
            ) : null}
          </div>

          <Separator />

          <section className="space-y-3" aria-labelledby="product-skus-title">
            <div className="space-y-1">
              <h2
                id="product-skus-title"
                className="font-heading text-lg font-medium"
              >
                {t("detail.skus.title")}
              </h2>
              <p className="text-sm text-muted-foreground">
                {t("detail.skus.description")}
              </p>
            </div>
            <ProductSKUPicker
              skus={product.skus}
              selectedSKUID={selectedSKU?.id ?? null}
              onSelect={onSelect}
            />
          </section>

          {added ? (
            <Alert>
              <CheckCircle2 className="text-emerald-600" aria-hidden />
              <AlertDescription>{t("detail.added")}</AlertDescription>
            </Alert>
          ) : null}
          {actionError ? (
            <Alert variant="destructive">
              <AlertDescription>
                {describeProductActionError(actionError, t)}
              </AlertDescription>
            </Alert>
          ) : null}

          <div className="grid gap-3 border-t pt-5 sm:grid-cols-[9rem_minmax(0,1fr)] sm:items-end">
            <div className="space-y-1.5">
              <p className="text-sm font-medium">{t("detail.quantity")}</p>
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="icon-sm"
                  aria-label={t("detail.decrease")}
                  disabled={!purchasable || isAdding || quantity <= 1}
                  onClick={() => onQuantityChange(quantity - 1)}
                >
                  −
                </Button>
                <Input
                  type="number"
                  min={1}
                  max={maxQuantity}
                  value={quantity}
                  aria-label={t("detail.quantity")}
                  disabled={!purchasable || isAdding}
                  className="h-8 w-16 text-center tabular-nums"
                  onChange={(event) => {
                    const next = Number(event.target.value)
                    if (Number.isInteger(next) && next >= 1) {
                      onQuantityChange(Math.min(next, maxQuantity))
                    }
                  }}
                />
                <Button
                  variant="outline"
                  size="icon-sm"
                  aria-label={t("detail.increase")}
                  disabled={
                    !purchasable || isAdding || quantity >= maxQuantity
                  }
                  onClick={() => onQuantityChange(quantity + 1)}
                >
                  +
                </Button>
              </div>
            </div>

            <div className="grid gap-2 sm:grid-cols-2">
              <Button
                variant="outline"
                disabled={!selectedSKU || !purchasable || isAdding}
                onClick={() => {
                  if (selectedSKU) {
                    void onAddToCart(selectedSKU)
                  }
                }}
              >
                {isAdding ? (
                  <LoaderCircle className="animate-spin" aria-hidden />
                ) : (
                  <ShoppingCart aria-hidden />
                )}
                {t("detail.addToCart")}
              </Button>
              {directCheckoutHref ? (
                <Link
                  to={directCheckoutHref}
                  className={buttonVariants()}
                >
                  <Zap aria-hidden />
                  {t("detail.buyNow")}
                </Link>
              ) : (
                <Button disabled>
                  <Zap aria-hidden />
                  {t("detail.buyNow")}
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function ProductDetailSkeleton() {
  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
      <Skeleton className="aspect-[4/5] rounded-xl" />
      <div className="space-y-5">
        <Skeleton className="h-5 w-28" />
        <Skeleton className="h-9 w-3/4" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-2/3" />
        <Skeleton className="h-8 w-36" />
        <Separator />
        <Skeleton className="h-5 w-40" />
        <div className="space-y-2">
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
        </div>
      </div>
    </div>
  )
}

export function ProductSignedOut() {
  const { t } = useTranslation("feed")

  return (
    <Card className="items-center gap-4 px-6 py-20 text-center">
      <div className="flex size-12 items-center justify-center rounded-full bg-muted">
        <Sparkles className="size-6 text-muted-foreground" aria-hidden />
      </div>
      <CardHeader className="items-center gap-1 px-0">
        <CardTitle className="text-2xl font-semibold">
          {t("productSignedOut.title")}
        </CardTitle>
        <CardDescription className="max-w-md text-center">
          {t("productSignedOut.description")}
        </CardDescription>
      </CardHeader>
      <CardContent className="p-0">
        <Link to="/signin" className={buttonVariants()}>
          {t("productSignedOut.action")}
        </Link>
      </CardContent>
    </Card>
  )
}
