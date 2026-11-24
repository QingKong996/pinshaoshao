import { useQueryClient } from "@tanstack/react-query"
import { useAuth } from "@/features/auth"
import {
  OrderDetailPage,
  OrderSignedOut,
  orderQueryKeys,
} from "@/features/order"
import { OrderPaymentPanel } from "@/features/payment"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"
import { useParams } from "react-router"

export function OrderDetailRoute() {
  const { session } = useAuth()
  const { orderID } = useParams()
  const queryClient = useQueryClient()

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        {session ? (
          <OrderDetailPage
            token={session.token}
            orderID={Number(orderID)}
            renderSidebar={(order) => {
              if (
                order.status !== "pending_payment" &&
                order.status !== "paid"
              ) {
                return null
              }
              return (
                <OrderPaymentPanel
                  token={session.token}
                  orderID={order.id}
                  orderStatus={order.status}
                  amountCents={order.total_amount_cents}
                  expiresAt={order.expires_at}
                  paidAt={order.paid_at}
                  onPaymentSettled={() => {
                    void queryClient.invalidateQueries({
                      queryKey: orderQueryKeys.order(
                        session.token,
                        order.id,
                      ),
                    })
                  }}
                />
              )
            }}
          />
        ) : (
          <OrderSignedOut />
        )}
      </main>
      <SiteFooter />
    </div>
  )
}
