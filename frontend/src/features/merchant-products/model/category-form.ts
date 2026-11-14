import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"
import type { ProductCategory, ProductCategoryInput } from "../types"

const CATEGORY_NAME_MAX = 50

const VALIDATION_MESSAGES = {
  nameRequired: "categoryForm.errors.nameRequired",
  nameLength: "categoryForm.errors.nameLength",
} as const

export const categoryFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.nameRequired)
    .max(CATEGORY_NAME_MAX, VALIDATION_MESSAGES.nameLength),
})

export type FormValues = z.infer<typeof categoryFormSchema>

export const INITIAL_VALUES: FormValues = { name: "" }

export function categoryToValues(category: ProductCategory): FormValues {
  return { name: category.name }
}

export function categoryToInput(values: FormValues): ProductCategoryInput {
  return { name: values.name }
}

export function translateValidationMessage(
  message: string | undefined,
  translate: TFunction<"console">,
) {
  if (!message) {
    return undefined
  }
  return message === VALIDATION_MESSAGES.nameLength
    ? translate(message as ParseKeys<"console">, { max: CATEGORY_NAME_MAX })
    : translate(message as ParseKeys<"console">)
}
