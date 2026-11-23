import { useInfiniteQuery } from "@tanstack/react-query"
import { listFeed } from "../api/feed"
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
