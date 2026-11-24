import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"

const MAX_SUB_ORDER_ID = 2147483647

const VALIDATION_MESSAGES = {
  subOrderRequired: "dispatchForm.errors.subOrderRequired",
  subOrderInvalid: "dispatchForm.errors.subOrderInvalid",
  carrierLength: "dispatchForm.errors.carrierLength",
} as const

export const dispatchFormSchema = z.object({
  subOrderID: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.subOrderRequired)
    .regex(/^[1-9]\d*$/, VALIDATION_MESSAGES.subOrderInvalid)
    .refine(
      (value) => Number(value) <= MAX_SUB_ORDER_ID,
      VALIDATION_MESSAGES.subOrderInvalid,
    ),
  carrier: z
    .string()
    .trim()
    .max(64, VALIDATION_MESSAGES.carrierLength),
})

export type DispatchFormValues = z.infer<typeof dispatchFormSchema>

export function translateDispatchValidation(
  message: string | undefined,
  translate: TFunction<"logistics">,
) {
  if (!message) {
    return undefined
  }
  return translate(message as ParseKeys<"logistics">)
}
