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
import {
  createProductCategory,
  deleteProductCategory,
  updateProductCategory,
} from "../api/categories"
import {
  createProductSKU,
  deleteProductSKU,
  updateProductSKU,
  updateProductSKUStock,
} from "../api/skus"
import type {
  CatalogChanged,
  ProductCategoryInput,
  ProductInput,
  ProductSKUInput,
} from "../types"
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

type UpdateCategoryVariables = {
  categoryID: number
  input: ProductCategoryInput
}

type UpdateInventoryStockVariables = {
  productID: number
  skuID: number
  stock: number
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
    queryClient.invalidateQueries({
      queryKey: productQueryKeys.inventory(),
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
    queryClient.invalidateQueries({
      queryKey: productQueryKeys.inventory(),
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

export function useCreateProductCategory(
  token: string,
  onCatalogChanged: CatalogChanged,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: ProductCategoryInput) =>
      createProductCategory(token, input),
    onSuccess: () => invalidateProductData(queryClient, onCatalogChanged),
  })
}

export function useUpdateProductCategory(
  token: string,
  onCatalogChanged: CatalogChanged,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ categoryID, input }: UpdateCategoryVariables) =>
      updateProductCategory(token, categoryID, input),
    onSuccess: () => invalidateProductData(queryClient, onCatalogChanged),
  })
}

export function useDeleteProductCategory(
  token: string,
  onCatalogChanged: CatalogChanged,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (categoryID: number) =>
      deleteProductCategory(token, categoryID),
    onSuccess: () => invalidateProductData(queryClient, onCatalogChanged),
  })
}

export function useUpdateInventoryStock(
  token: string,
  onCatalogChanged: CatalogChanged,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      productID,
      skuID,
      stock,
    }: UpdateInventoryStockVariables) =>
      updateProductSKUStock(token, productID, skuID, stock),
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({
          queryKey: productQueryKeys.inventory(),
        }),
        queryClient.invalidateQueries({
          queryKey: productQueryKeys.products(),
        }),
        onCatalogChanged(),
      ]),
  })
}
