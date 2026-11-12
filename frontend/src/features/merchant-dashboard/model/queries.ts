import { useQuery, type QueryClient } from "@tanstack/react-query"
import { getDashboard } from "../api/dashboard"
import { getShopProfile } from "../api/shop-profile"
import { dashboardQueryKeys } from "./query-keys"

export function useDashboard(token: string) {
  return useQuery({
    queryKey: dashboardQueryKeys.dashboard(),
    queryFn: () => getDashboard(token),
  })
}

export function useShopProfile(token: string) {
  return useQuery({
    queryKey: dashboardQueryKeys.shop(),
    queryFn: () => getShopProfile(token),
  })
}

export function invalidateMerchantDashboard(queryClient: QueryClient) {
  return queryClient.invalidateQueries({
    queryKey: dashboardQueryKeys.dashboard(),
  })
}
