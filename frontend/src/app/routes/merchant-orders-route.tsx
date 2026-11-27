import { useQueryClient } from "@tanstack/react-query"
import { useMerchantSession } from "@/features/auth"
import {
  MerchantOrdersPage,
  merchantOrderQueryKeys,
} from "@/features/merchant-orders"
import { ShipSubOrderButton } from "@/features/logistics"

export function MerchantOrdersRoute() {
  const { token } = useMerchantSession()
  const queryClient = useQueryClient()

  return (
    <MerchantOrdersPage
      token={token}
      renderRowActions={(order) =>
        order.status === "awaiting_shipment" ? (
          <ShipSubOrderButton
            token={token}
            subOrderID={order.id}
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
