import { useMerchantSession } from "@/features/auth"
import { MerchantDashboardPage } from "@/features/merchant-dashboard"

export function MerchantDashboardRoute() {
  const { token } = useMerchantSession()
  return <MerchantDashboardPage token={token} />
}
