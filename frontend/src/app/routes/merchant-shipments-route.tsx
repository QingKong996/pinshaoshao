import { useMerchantSession } from "@/features/auth"
import { MerchantShipmentsPage } from "@/features/logistics"

export function MerchantShipmentsRoute() {
  const { token } = useMerchantSession()
  return <MerchantShipmentsPage token={token} />
}
