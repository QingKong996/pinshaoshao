import { useQueryClient } from "@tanstack/react-query"
import { useParams } from "react-router"
import { useMerchantSession } from "@/features/auth"
import {
  MerchantOrderDetailPage,
  merchantOrderQueryKeys,
} from "@/features/merchant-orders"
import { ShipSubOrderButton } from "@/features/logistics"
import { MERCHANT_CONSOLE_PATHS } from "../merchant-navigation"

export function MerchantOrderDetailRoute() {
  const { token } = useMerchantSession()
  const { subOrderID } = useParams()
  const queryClient = useQueryClient()

  return (
    <MerchantOrderDetailPage
      token={token}
      subOrderID={Number(subOrderID)}
      backHref={MERCHANT_CONSOLE_PATHS.orders}
      renderActions={(order) =>
        order.status === "awaiting_shipment" ? (
          <ShipSubOrderButton
            token={token}
            subOrderID={order.id}
            receiverName={order.receiver_name}
            onCreated={() => {
              void queryClient.invalidateQueries({
                queryKey: merchantOrderQueryKeys.orders(),
              })
            }}
          />
        ) : null
      }
    />
  )
}
