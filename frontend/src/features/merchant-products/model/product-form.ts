import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"
import type { Product, ProductInput } from "../types"

const PRODUCT_NAME_MAX = 100
export const UNCATEGORIZED = "none"

const VALIDATION_MESSAGES = {
  nameRequired: "productForm.errors.nameRequired",
  nameLength: "productForm.errors.nameLength",
} as const

export const productFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.nameRequired)
    .max(PRODUCT_NAME_MAX, VALIDATION_MESSAGES.nameLength),
  description: z.string(),
  categoryId: z.string(),
  status: z.enum(["on_shelf", "off_shelf"]),
})

export type FormValues = z.infer<typeof productFormSchema>

export const INITIAL_VALUES: FormValues = {
  name: "",
  description: "",
  categoryId: UNCATEGORIZED,
  status: "off_shelf",
}

export function productToValues(product: Product): FormValues {
  return {
    name: product.name,
    description: product.description,
    categoryId:
      product.category_id === null ? UNCATEGORIZED : String(product.category_id),
    status: product.status,
  }
}

export function productToInput(values: FormValues): ProductInput {
  return {
    name: values.name,
    description: values.description,
    category_id:
      values.categoryId === UNCATEGORIZED ? null : Number(values.categoryId),
    status: values.status,
  }
}

export function translateValidationMessage(
  message: string | undefined,
  translate: TFunction<"console">,
) {
  if (!message) {
    return undefined
  }
  return message === VALIDATION_MESSAGES.nameLength
    ? translate(message as ParseKeys<"console">, { max: PRODUCT_NAME_MAX })
    : translate(message as ParseKeys<"console">)
}
