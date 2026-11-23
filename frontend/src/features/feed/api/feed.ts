import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { FeedListResponse, ProductDetail } from "../types"

const FEED_PAGE_SIZE = 20

export async function listFeed(token: string, page: number) {
  const { data } = await apiClient.get<FeedListResponse>("/feed", {
    ...authorized(token),
    params: { page, page_size: FEED_PAGE_SIZE },
  })
  return data
}

export async function getProductDetail(token: string, productID: number) {
  const { data } = await apiClient.get<ProductDetail>(
    `/products/${productID}`,
    authorized(token),
  )
  return data
}
