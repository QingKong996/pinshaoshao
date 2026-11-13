import { useQueryClient } from "@tanstack/react-query"
import { useParams } from "react-router"
import { useMerchantSession } from "@/features/auth"
import { MerchantProductDetailPage } from "@/features/merchant-products"
import { invalidateMerchantDashboard } from "@/features/merchant-dashboard"
import { MERCHANT_CONSOLE_PATHS } from "../merchant-navigation"

export function MerchantProductDetailRoute() {
  const { token } = useMerchantSession()
  const { productID } = useParams()
  const queryClient = useQueryClient()
  return (
    <MerchantProductDetailPage
      token={token}
      productID={Number(productID)}
      backHref={MERCHANT_CONSOLE_PATHS.products}
      onCatalogChanged={() => invalidateMerchantDashboard(queryClient)}
    />
  )
}
