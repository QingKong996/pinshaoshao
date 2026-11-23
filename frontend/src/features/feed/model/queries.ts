import { useInfiniteQuery, useQuery } from "@tanstack/react-query"
import { getProductDetail, listFeed } from "../api/feed"
import { feedQueryKeys } from "./query-keys"

export function useFeed(token: string) {
  return useInfiniteQuery({
    queryKey: feedQueryKeys.feed(token),
    queryFn: ({ pageParam }) => listFeed(token, pageParam),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.page < lastPage.total_pages ? lastPage.page + 1 : undefined,
    enabled: token.length > 0,
  })
}

export function useProductDetail(
  token: string,
  productID: number,
  enabled = true,
) {
  return useQuery({
    queryKey: feedQueryKeys.product(token, productID),
    queryFn: () => getProductDetail(token, productID),
    enabled: enabled && token.length > 0 && productID > 0,
  })
}
