import { useQueryClient } from "@tanstack/react-query"
import { useMerchantSession } from "@/features/auth"
import { MerchantCategoriesPage } from "@/features/merchant-products"
import { invalidateMerchantDashboard } from "@/features/merchant-dashboard"

export function MerchantCategoriesRoute() {
  const { token } = useMerchantSession()
  const queryClient = useQueryClient()
  return (
    <MerchantCategoriesPage
      token={token}
      onCatalogChanged={() => invalidateMerchantDashboard(queryClient)}
    />
  )
}
