import {
  useMutation,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query"
import {
  createProduct,
  createProductSKU,
  deleteProduct,
  deleteProductSKU,
  updateProduct,
  updateProductSKU,
  type ProductInput,
  type ProductSKUInput,
} from "@/api/merchant-console"
import { merchantConsoleQueryKeys } from "@/merchant-console/query-keys"

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

function invalidateProductData(queryClient: QueryClient) {
  return Promise.all([
    // Prefix match also refreshes product detail and SKU queries.
    queryClient.invalidateQueries({
      queryKey: merchantConsoleQueryKeys.products(),
    }),
    queryClient.invalidateQueries({
      queryKey: merchantConsoleQueryKeys.categories(),
    }),
    queryClient.invalidateQueries({
      queryKey: merchantConsoleQueryKeys.dashboard(),
    }),
  ])
}

function invalidateSKUData(queryClient: QueryClient) {
  return Promise.all([
    queryClient.invalidateQueries({
      queryKey: merchantConsoleQueryKeys.products(),
    }),
    queryClient.invalidateQueries({
      queryKey: merchantConsoleQueryKeys.dashboard(),
    }),
  ])
}

export function useCreateProduct(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: ProductInput) => createProduct(token, input),
    onSuccess: () => invalidateProductData(queryClient),
  })
}

export function useUpdateProduct(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productID, input }: UpdateProductVariables) =>
      updateProduct(token, productID, input),
    onSuccess: () => invalidateProductData(queryClient),
  })
}

export function useDeleteProduct(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (productID: number) => deleteProduct(token, productID),
    onSuccess: () => invalidateProductData(queryClient),
  })
}

export function useCreateProductSKU(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productID, input }: CreateProductSKUVariables) =>
      createProductSKU(token, productID, input),
    onSuccess: () => invalidateSKUData(queryClient),
  })
}

export function useUpdateProductSKU(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productID, skuID, input }: UpdateProductSKUVariables) =>
      updateProductSKU(token, productID, skuID, input),
    onSuccess: () => invalidateSKUData(queryClient),
  })
}

export function useDeleteProductSKU(token: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ productID, skuID }: DeleteProductSKUVariables) =>
      deleteProductSKU(token, productID, skuID),
    onSuccess: () => invalidateSKUData(queryClient),
  })
}
