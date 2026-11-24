import { useMerchantSession } from "@/features/auth"
import { MerchantOrdersPage } from "@/features/merchant-orders"

export function MerchantOrdersRoute() {
  const { token } = useMerchantSession()
  return <MerchantOrdersPage token={token} />
}
