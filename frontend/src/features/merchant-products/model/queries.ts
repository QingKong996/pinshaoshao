import { useQuery } from "@tanstack/react-query"
import { getProduct, listProducts } from "../api/products"
import { listProductCategories } from "../api/categories"
import { listProductSKUs } from "../api/skus"
import { productQueryKeys } from "./query-keys"

export function useProducts(token: string) {
  return useQuery({
    queryKey: productQueryKeys.products(),
    queryFn: () => listProducts(token),
  })
}

export function useProduct(token: string, productID: number) {
  return useQuery({
    queryKey: productQueryKeys.product(productID),
    queryFn: () => getProduct(token, productID),
    enabled: Number.isInteger(productID) && productID > 0,
  })
}

export function useProductCategories(token: string, enabled = true) {
  return useQuery({
    queryKey: productQueryKeys.categories(),
    queryFn: () => listProductCategories(token),
    enabled,
  })
}

export function useProductSKUs(token: string, productID: number) {
  return useQuery({
    queryKey: productQueryKeys.productSKUs(productID),
    queryFn: () => listProductSKUs(token, productID),
    enabled: Number.isInteger(productID) && productID > 0,
  })
}
