import { useQueryClient } from "@tanstack/react-query"
import { useMerchantSession } from "@/features/auth"
import { MerchantInventoryPage } from "@/features/merchant-products"
import { invalidateMerchantDashboard } from "@/features/merchant-dashboard"
import { merchantConsoleProductPath } from "../merchant-navigation"

export function MerchantInventoryRoute() {
  const { token } = useMerchantSession()
  const queryClient = useQueryClient()
  return (
    <MerchantInventoryPage
      token={token}
      productHref={merchantConsoleProductPath}
      onCatalogChanged={() => invalidateMerchantDashboard(queryClient)}
    />
  )
}
