import { useMemo } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { Link, useSearchParams } from "react-router"
import { useAuth } from "@/features/auth"
import { useAddresses } from "@/features/address"
import {
  cartQueryKeys,
  describeCartError,
  useCart,
} from "@/features/cart"
import {
  describeProductError,
  feedQueryKeys,
  isSKUPurchasable,
  useProductDetail,
} from "@/features/feed"
import {
  CheckoutPage,
  OrderSignedOut,
  type CheckoutLine,
} from "@/features/order"
import { buttonVariants } from "@/shared/ui/button"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"

export function CheckoutRoute() {
  const { session } = useAuth()
  const { t } = useTranslation("order")
  const { t: cartT } = useTranslation("cart")
  const { t: feedT } = useTranslation("feed")
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()

  const isDirect = ["product_id", "sku_id", "quantity"].some((key) =>
    searchParams.has(key),
  )
  const directProductID = positiveInteger(searchParams.get("product_id"))
  const directSKUID = positiveInteger(searchParams.get("sku_id"))
  const directQuantity = positiveInteger(searchParams.get("quantity") ?? "1")
  const validDirectLink =
    directProductID !== null &&
    directSKUID !== null &&
    directQuantity !== null

  const cartQuery = useCart(session?.token ?? "", !isDirect)
  const addressesQuery = useAddresses(session?.token ?? "")
  const productQuery = useProductDetail(
    session?.token ?? "",
    directProductID ?? 0,
    isDirect && validDirectLink,
  )
  const selectedParam = searchParams.get("items")
  const selectedIDs = useMemo(() => {
    if (selectedParam === null) {
      return null
    }
    return new Set(
      selectedParam
        .split(",")
        .map((value) => Number(value))
        .filter((value) => Number.isInteger(value) && value > 0),
    )
  }, [selectedParam])

  const cartItems = (cartQuery.data ?? [])
    .filter(
      (item) =>
        item.availability === "available" &&
        (selectedIDs === null || selectedIDs.has(item.id)),
    )
    .map((item) => ({
      id: item.id,
      product_name: item.product_name,
      sku_specs: item.sku_specs,
      unit_price_cents: item.unit_price_cents,
      quantity: item.quantity,
      line_total_cents: item.line_total_cents,
    }))

  let content
  if (!session) {
    content = <OrderSignedOut />
  } else if (isDirect) {
    if (
      !validDirectLink ||
      !directProductID ||
      !directSKUID ||
      !directQuantity
    ) {
      const backHref = directProductID ? `/products/${directProductID}` : "/"
      content = (
        <QueryError
          title={t("checkout.direct.invalidTitle")}
          description={t("checkout.direct.invalidDescription")}
          action={
            <Link
              to={backHref}
              className={buttonVariants({ variant: "outline", size: "sm" })}
            >
              {directProductID
                ? t("checkout.backToProduct")
                : feedT("detail.back")}
            </Link>
          }
        />
      )
    } else if (productQuery.error) {
      content = (
        <QueryError
          title={feedT("errors.productLoadTitle")}
          description={describeProductError(productQuery.error, feedT)}
          retryLabel={feedT("errors.retry")}
          onRetry={productQuery.refetch}
        />
      )
    } else if (productQuery.isPending || !productQuery.data) {
      content = <Skeleton className="h-96 rounded-xl" />
    } else {
      const sku = productQuery.data.skus.find(
        (candidate) => candidate.id === directSKUID,
      )

      if (!sku || !isSKUPurchasable(sku)) {
        content = (
          <QueryError
            title={t("checkout.direct.unavailableTitle")}
            description={t("checkout.direct.unavailableDescription")}
            action={
              <Link
                to={`/products/${directProductID}`}
                className={buttonVariants({
                  variant: "outline",
                  size: "sm",
                })}
              >
                {t("checkout.backToProduct")}
              </Link>
            }
          />
        )
      } else if (addressesQuery.isPending) {
        content = <Skeleton className="h-96 rounded-xl" />
      } else {
        const items: CheckoutLine[] = [
          {
            id: sku.id,
            product_name: productQuery.data.name,
            sku_specs: sku.specs,
            unit_price_cents: sku.price_cents,
            quantity: directQuantity,
            line_total_cents: sku.price_cents * directQuantity,
          },
        ]
        content = (
          <CheckoutPage
            token={session.token}
            items={items}
            savedAddresses={addressesQuery.data}
            mode={{ type: "direct", skuID: directSKUID }}
            backHref={`/products/${directProductID}`}
            onOrderCreated={() => {
              void queryClient.invalidateQueries({
                queryKey: feedQueryKeys.feed(session.token),
              })
            }}
          />
        )
      }
    }
  } else if (cartQuery.error && !cartQuery.data) {
    content = (
      <QueryError
        title={cartT("errors.loadTitle")}
        description={describeCartError(cartQuery.error, cartT)}
        retryLabel={cartT("errors.retry")}
        onRetry={cartQuery.refetch}
      />
    )
  } else if (cartQuery.isPending) {
    content = <Skeleton className="h-96 rounded-xl" />
  } else if (addressesQuery.isPending) {
    content = <Skeleton className="h-96 rounded-xl" />
  } else {
    content = (
      <CheckoutPage
        token={session.token}
        items={cartItems}
        savedAddresses={addressesQuery.data}
        onOrderCreated={() => {
          void queryClient.invalidateQueries({
            queryKey: cartQueryKeys.cart(session.token),
          })
        }}
      />
    )
  }

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        {content}
      </main>
      <SiteFooter />
    </div>
  )
}

function positiveInteger(value: string | null) {
  if (value === null) {
    return null
  }
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null
}
