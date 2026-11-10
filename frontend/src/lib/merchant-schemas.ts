import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"

const SHOP_NAME_MAX = 50
const PRODUCT_NAME_MAX = 100
const CATEGORY_NAME_MAX = 50
const SHOP_DESCRIPTION_MAX = 500
const PRODUCT_DESCRIPTION_MAX = 2000
const SPEC_KEY_MAX = 20
const SPEC_VALUE_MAX = 50
export const SPEC_MAX_COUNT = 5
const STOCK_MAX = 2147483647
const PRICE_MAX = 1000000
const PRICE_PATTERN = /^\d+(\.\d{1,2})?$/
const INTEGER_PATTERN = /^\d+$/

const VALIDATION_KEYS = {
  nameRequired: "validation.nameRequired",
  nameTooLong: "validation.nameTooLong",
  productNameTooLong: "validation.productNameTooLong",
  categoryNameTooLong: "validation.categoryNameTooLong",
  descriptionTooLong: "validation.descriptionTooLong",
  productDescriptionTooLong: "validation.productDescriptionTooLong",
  priceRequired: "validation.priceRequired",
  priceInvalid: "validation.priceInvalid",
  priceTooLarge: "validation.priceTooLarge",
  stockRequired: "validation.stockRequired",
  stockInvalid: "validation.stockInvalid",
  stockTooLarge: "validation.stockTooLarge",
  specRequired: "validation.specRequired",
  specKeyRequired: "validation.specKeyRequired",
  specValueRequired: "validation.specValueRequired",
  specKeyTooLong: "validation.specKeyTooLong",
  specValueTooLong: "validation.specValueTooLong",
  specKeyDuplicate: "validation.specKeyDuplicate",
  specsTooMany: "validation.specsTooMany",
} as const satisfies Record<string, ParseKeys<"merchant">>

const MESSAGE_VALUES: Partial<
  Record<ParseKeys<"merchant">, Record<string, number>>
> = {
  [VALIDATION_KEYS.nameTooLong]: { max: SHOP_NAME_MAX },
  [VALIDATION_KEYS.productNameTooLong]: { max: PRODUCT_NAME_MAX },
  [VALIDATION_KEYS.categoryNameTooLong]: { max: CATEGORY_NAME_MAX },
  [VALIDATION_KEYS.descriptionTooLong]: { max: SHOP_DESCRIPTION_MAX },
  [VALIDATION_KEYS.productDescriptionTooLong]: {
    max: PRODUCT_DESCRIPTION_MAX,
  },
  [VALIDATION_KEYS.specKeyTooLong]: { max: SPEC_KEY_MAX },
  [VALIDATION_KEYS.specValueTooLong]: { max: SPEC_VALUE_MAX },
  [VALIDATION_KEYS.specsTooMany]: { max: SPEC_MAX_COUNT },
}

export function translateMerchantValidationMessage(
  message: string | undefined,
  t: TFunction<"merchant">,
): string | undefined {
  if (!message) {
    return undefined
  }
  const key = message as ParseKeys<"merchant">
  const values = MESSAGE_VALUES[key]
  return values ? t(key, values) : t(key)
}

const nameSchema = z
  .string()
  .trim()
  .min(1, VALIDATION_KEYS.nameRequired)
  .max(SHOP_NAME_MAX, VALIDATION_KEYS.nameTooLong)

const priceSchema = z
  .string()
  .trim()
  .min(1, VALIDATION_KEYS.priceRequired)
  .refine((value) => PRICE_PATTERN.test(value), VALIDATION_KEYS.priceInvalid)
  .refine(
    (value) => PRICE_PATTERN.test(value) && Number(value) <= PRICE_MAX,
    VALIDATION_KEYS.priceTooLarge,
  )

const stockSchema = z
  .string()
  .trim()
  .min(1, VALIDATION_KEYS.stockRequired)
  .refine((value) => INTEGER_PATTERN.test(value), VALIDATION_KEYS.stockInvalid)
  .refine(
    (value) => INTEGER_PATTERN.test(value) && Number(value) <= STOCK_MAX,
    VALIDATION_KEYS.stockTooLarge,
  )

export const merchantShopProfileSchema = z.object({
  name: nameSchema,
  description: z
    .string()
    .trim()
    .max(SHOP_DESCRIPTION_MAX, VALIDATION_KEYS.descriptionTooLong),
})

export type MerchantShopProfileValues = z.infer<
  typeof merchantShopProfileSchema
>

export const merchantCategorySchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, VALIDATION_KEYS.nameRequired)
    .max(CATEGORY_NAME_MAX, VALIDATION_KEYS.categoryNameTooLong),
})

export type MerchantCategoryValues = z.infer<typeof merchantCategorySchema>

export const merchantProductSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, VALIDATION_KEYS.nameRequired)
    .max(PRODUCT_NAME_MAX, VALIDATION_KEYS.productNameTooLong),
  description: z
    .string()
    .trim()
    .max(PRODUCT_DESCRIPTION_MAX, VALIDATION_KEYS.productDescriptionTooLong),
  categoryId: z.string(),
  status: z.enum(["on_shelf", "off_shelf"]),
})

export type MerchantProductValues = z.infer<typeof merchantProductSchema>

export const merchantSKUSchema = z
  .object({
    specs: z
      .array(
        z.object({
          key: z
            .string()
            .trim()
            .min(1, VALIDATION_KEYS.specKeyRequired)
            .max(SPEC_KEY_MAX, VALIDATION_KEYS.specKeyTooLong),
          value: z
            .string()
            .trim()
            .min(1, VALIDATION_KEYS.specValueRequired)
            .max(SPEC_VALUE_MAX, VALIDATION_KEYS.specValueTooLong),
        }),
      )
      .min(1, VALIDATION_KEYS.specRequired)
      .max(SPEC_MAX_COUNT, VALIDATION_KEYS.specsTooMany),
    price: priceSchema,
    stock: stockSchema,
    isEnabled: z.boolean(),
  })
  .superRefine((values, ctx) => {
    const seen = new Set<string>()
    for (const spec of values.specs) {
      const key = spec.key.trim()
      if (!key) {
        continue
      }
      if (seen.has(key)) {
        ctx.addIssue({
          code: "custom",
          message: VALIDATION_KEYS.specKeyDuplicate,
          path: ["specs"],
        })
        return
      }
      seen.add(key)
    }
  })

export type MerchantSKUValues = z.infer<typeof merchantSKUSchema>

export const merchantStockSchema = z.object({
  stock: stockSchema,
})

export type MerchantStockValues = z.infer<typeof merchantStockSchema>
