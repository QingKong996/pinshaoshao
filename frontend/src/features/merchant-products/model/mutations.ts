import {
  useMutation,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query"
import {
  createProduct,
  deleteProduct,
  updateProduct,
} from "../api/products"
import { createProductSKU, deleteProductSKU, updateProductSKU } from "../api/skus"
import type { ProductInput, ProductSKUInput, CatalogChanged } from "../types"
import { productQueryKeys } from "./query-keys"

type UpdateProductVariables = {
  productID: number
  input: ProductInput
}

type CreateProductSKUVariables = {
  productID: number
  input: ProductSKUInput
}

type UpdateProductSKUVariables = CreateProductSKUVariables & {
  skuID: number
}

type DeleteProductSKUVariables = {
  productID: number
  skuID: number
}

function invalidateProductData(
  queryClient: QueryClient,
  onCatalogChanged: CatalogChanged,
) {
  return Promise.all([
    // Prefix match also refreshes product detail and SKU queries.
    queryClient.invalidateQueries({
      queryKey: productQueryKeys.products(),
    }),
    queryClient.invalidateQueries({
      queryKey: productQueryKeys.categories(),
    }),
    onCatalogChanged(),
  ])
}

function invalidateSKUData(
  queryClient: QueryClient,
  onCatalogChanged: CatalogChanged,
) {
  return Promise.all([
    queryClient.invalidateQueries({
      queryKey: productQueryKeys.products(),
    }),
    onCatalogChanged(),
  ])
}

export function useCreateProduct(token: string, onCatalogChanged: CatalogChanged) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: ProductInput) => createProduct(token, input),
    onSuccess: () => invalidateProductData(queryClient, onCatalogChanged),
  })
}

export function useUpdateProduct(token: string, onCatalogChanged: CatalogChanged) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productID, input }: UpdateProductVariables) =>
      updateProduct(token, productID, input),
    onSuccess: () => invalidateProductData(queryClient, onCatalogChanged),
  })
}

export function useDeleteProduct(token: string, onCatalogChanged: CatalogChanged) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (productID: number) => deleteProduct(token, productID),
    onSuccess: () => invalidateProductData(queryClient, onCatalogChanged),
  })
}

export function useCreateProductSKU(
  token: string,
  onCatalogChanged: CatalogChanged,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productID, input }: CreateProductSKUVariables) =>
      createProductSKU(token, productID, input),
    onSuccess: () => invalidateSKUData(queryClient, onCatalogChanged),
  })
}

export function useUpdateProductSKU(
  token: string,
  onCatalogChanged: CatalogChanged,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productID, skuID, input }: UpdateProductSKUVariables) =>
      updateProductSKU(token, productID, skuID, input),
    onSuccess: () => invalidateSKUData(queryClient, onCatalogChanged),
  })
}

export function useDeleteProductSKU(
  token: string,
  onCatalogChanged: CatalogChanged,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productID, skuID }: DeleteProductSKUVariables) =>
      deleteProductSKU(token, productID, skuID),
    onSuccess: () => invalidateSKUData(queryClient, onCatalogChanged),
  })
}
