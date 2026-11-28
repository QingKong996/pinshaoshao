import { useMemo, useState, type ComponentProps } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { Link, useLocation, useSearchParams } from "react-router"
import { useAuth } from "@/features/auth"
import { useAddresses } from "@/features/address"
import {
  cartQueryKeys,
  cartSelectionSnapshot,
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
  cartCheckoutAttemptKey,
  OrderSignedOut,
  directCheckoutAttemptKey,
  readDirectCheckoutAttempt,
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
  const location = useLocation()
  const [searchParams] = useSearchParams()
  const cartSelectionBaseline = readCartSelectionSnapshot(location.state)

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
    } else if (productQuery.isPending) {
      content = <Skeleton className="h-96 rounded-xl" />
    } else {
      const product = productQuery.data
      const sku = product?.skus.find(
        (candidate) => candidate.id === directSKUID,
      )
      const attemptStorageKey = directCheckoutAttemptKey(
        session.username,
        directSKUID,
        directQuantity,
      )
      const hasPendingAttempt =
        readDirectCheckoutAttempt(attemptStorageKey) !== null
      const canStartNew =
        !productQuery.error && sku !== undefined && isSKUPurchasable(sku)

      if (productQuery.error && !hasPendingAttempt) {
        content = (
          <QueryError
            title={feedT("errors.productLoadTitle")}
            description={describeProductError(productQuery.error, feedT)}
            retryLabel={feedT("errors.retry")}
            onRetry={productQuery.refetch}
          />
        )
      } else if (!canStartNew && !hasPendingAttempt) {
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
        const items: CheckoutLine[] = product && sku
          ? [
              {
                id: sku.id,
                product_name: product.name,
                sku_specs: sku.specs,
                unit_price_cents: sku.price_cents,
                quantity: directQuantity,
                line_total_cents: sku.price_cents * directQuantity,
              },
            ]
          : []
        content = (
          <CheckoutPage
            key={attemptStorageKey}
            token={session.token}
            items={items}
            savedAddresses={addressesQuery.data}
            mode={{
              type: "direct",
              skuID: directSKUID,
              quantity: directQuantity,
              attemptStorageKey,
              canStartNew,
            }}
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
    const cartItems = (cartQuery.data ?? []).filter(
      (item) =>
        item.availability === "available" &&
        (selectedIDs === null || selectedIDs.has(item.id)),
    )
    const cartAttemptStorageKey = cartCheckoutAttemptKey(session.username)
    const snapshot = cartSelectionSnapshot(cartItems)
    const initiallyChanged =
      (selectedIDs !== null && cartItems.length !== selectedIDs.size) ||
      (cartSelectionBaseline !== null && cartSelectionBaseline !== snapshot)

    content = (
      <CartCheckout
        key={`${location.key}:${session.token}:${cartAttemptStorageKey}`}
        token={session.token}
        items={cartItems}
        savedAddresses={addressesQuery.data}
        cartAttemptStorageKey={cartAttemptStorageKey}
        snapshot={snapshot}
        initiallyChanged={initiallyChanged}
        onCartConflict={async () => {
          await cartQuery.refetch()
        }}
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

function CartCheckout({
  snapshot,
  initiallyChanged,
  ...checkoutProps
}: ComponentProps<typeof CheckoutPage> & {
  snapshot: string
  initiallyChanged: boolean
}) {
  const [review, setReview] = useState(() => ({
    snapshot,
    needsConfirmation: initiallyChanged,
  }))
  if (review.snapshot !== snapshot) {
    setReview({ snapshot, needsConfirmation: true })
  }

  return (
    <CheckoutPage
      {...checkoutProps}
      cartChangeNeedsConfirmation={review.needsConfirmation}
      onConfirmCartChanges={() =>
        setReview({ snapshot, needsConfirmation: false })
      }
    />
  )
}

function readCartSelectionSnapshot(state: unknown): string | null {
  if (state === null || typeof state !== "object") {
    return null
  }
  const value = (state as { cartSelectionSnapshot?: unknown })
    .cartSelectionSnapshot
  return typeof value === "string" ? value : null
}

function positiveInteger(value: string | null) {
  if (value === null) {
    return null
  }
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null
}
