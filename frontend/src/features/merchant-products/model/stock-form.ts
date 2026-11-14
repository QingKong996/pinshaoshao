import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"

const STOCK_MAX = 2147483647

const VALIDATION_MESSAGES = {
  stockRequired: "stockForm.errors.stockRequired",
  stockInvalid: "stockForm.errors.stockInvalid",
} as const

export const stockFormSchema = z.object({
  stock: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.stockRequired)
    .regex(/^\d+$/, VALIDATION_MESSAGES.stockInvalid)
    .refine(
      (value) => Number(value) <= STOCK_MAX,
      VALIDATION_MESSAGES.stockInvalid,
    ),
})

export type FormValues = z.infer<typeof stockFormSchema>

export function initialValues(stock: number): FormValues {
  return { stock: String(stock) }
}

export function translateValidationMessage(
  message: string | undefined,
  translate: TFunction<"console">,
) {
  if (!message) {
    return undefined
  }
  return translate(message as ParseKeys<"console">)
}
