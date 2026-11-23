import { useAuth } from "@/features/auth"
import { OrderDetailPage, OrderSignedOut } from "@/features/order"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"
import { useParams } from "react-router"

export function OrderDetailRoute() {
  const { session } = useAuth()
  const { orderID } = useParams()

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        {session ? (
          <OrderDetailPage
            token={session.token}
            orderID={Number(orderID)}
          />
        ) : (
          <OrderSignedOut />
        )}
      </main>
      <SiteFooter />
    </div>
  )
}
