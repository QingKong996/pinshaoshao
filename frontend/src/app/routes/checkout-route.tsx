import { useMemo } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router"
import { useAuth } from "@/features/auth"
import {
  cartQueryKeys,
  describeCartError,
  useCart,
} from "@/features/cart"
import { CheckoutPage, OrderSignedOut } from "@/features/order"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"

export function CheckoutRoute() {
  const { session } = useAuth()
  const { t: cartT } = useTranslation("cart")
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const cartQuery = useCart(session?.token ?? "")
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

  const items = (cartQuery.data ?? [])
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

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        {!session ? (
          <OrderSignedOut />
        ) : cartQuery.error && !cartQuery.data ? (
          <QueryError
            title={cartT("errors.loadTitle")}
            description={describeCartError(cartQuery.error, cartT)}
            retryLabel={cartT("errors.retry")}
            onRetry={cartQuery.refetch}
          />
        ) : cartQuery.isPending ? (
          <Skeleton className="h-96 rounded-xl" />
        ) : (
          <CheckoutPage
            token={session.token}
            items={items}
            onOrderCreated={() => {
              void queryClient.invalidateQueries({
                queryKey: cartQueryKeys.cart(session.token),
              })
            }}
          />
        )}
      </main>
      <SiteFooter />
    </div>
  )
}
