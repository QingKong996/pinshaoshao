import { useQueryClient } from "@tanstack/react-query"
import { useMerchantSession } from "@/features/auth"
import { MerchantProductsPage } from "@/features/merchant-products"
import { invalidateMerchantDashboard } from "@/features/merchant-dashboard"
import { merchantConsoleProductPath } from "../merchant-navigation"

export function MerchantProductsRoute() {
  const { token } = useMerchantSession()
  const queryClient = useQueryClient()
  return (
    <MerchantProductsPage
      token={token}
      productHref={merchantConsoleProductPath}
      onCatalogChanged={() => invalidateMerchantDashboard(queryClient)}
    />
  )
}
