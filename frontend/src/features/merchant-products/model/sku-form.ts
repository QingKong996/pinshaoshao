import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"
import type { ProductSKU, ProductSKUInput } from "../types"

export const MAX_SPECS = 5
const PRICE_PATTERN = /^\d+(\.\d{1,2})?$/
const STOCK_PATTERN = /^\d+$/

const VALIDATION_MESSAGES = {
  specKeyRequired: "skuForm.errors.specKeyRequired",
  specValueRequired: "skuForm.errors.specValueRequired",
  specDuplicate: "skuForm.errors.specDuplicate",
  specLimit: "skuForm.errors.specLimit",
  priceRequired: "skuForm.errors.priceRequired",
  priceInvalid: "skuForm.errors.priceInvalid",
  stockRequired: "skuForm.errors.stockRequired",
  stockInvalid: "skuForm.errors.stockInvalid",
} as const

const skuFormSchema = z.object({
  specs: z.array(z.object({ key: z.string(), value: z.string() })),
  price: z.string(),
  stock: z.string(),
  is_enabled: z.boolean(),
})

export type FormValues = z.infer<typeof skuFormSchema>

export function buildFormSchema(specsEditable: boolean) {
  return skuFormSchema.superRefine((values, context) => {
    const seenKeys = new Set<string>()
    values.specs.forEach((spec, index) => {
      const key = spec.key.trim()
      if (specsEditable) {
        if (!key) {
          context.addIssue({
            code: "custom",
            message: VALIDATION_MESSAGES.specKeyRequired,
            path: ["specs", index, "key"],
          })
        } else if (seenKeys.has(key)) {
          context.addIssue({
            code: "custom",
            message: VALIDATION_MESSAGES.specDuplicate,
            path: ["specs", index, "key"],
          })
        }
        if (key) {
          seenKeys.add(key)
        }
      }
      if (!spec.value.trim()) {
        context.addIssue({
          code: "custom",
          message: VALIDATION_MESSAGES.specValueRequired,
          path: ["specs", index, "value"],
        })
      }
    })
    if (specsEditable && values.specs.length > MAX_SPECS) {
      context.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.specLimit,
        path: ["specs", MAX_SPECS],
      })
    }
    const price = values.price.trim()
    if (!price) {
      context.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.priceRequired,
        path: ["price"],
      })
    } else if (!PRICE_PATTERN.test(price)) {
      context.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.priceInvalid,
        path: ["price"],
      })
    }
    const stock = values.stock.trim()
    if (!stock) {
      context.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.stockRequired,
        path: ["stock"],
      })
    } else if (!STOCK_PATTERN.test(stock)) {
      context.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.stockInvalid,
        path: ["stock"],
      })
    }
  })
}

export function translateValidationMessage(
  message: string | undefined,
  translate: TFunction<"console">,
) {
  if (!message) {
    return undefined
  }
  return message === VALIDATION_MESSAGES.specLimit
    ? translate(message as ParseKeys<"console">, { max: MAX_SPECS })
    : translate(message as ParseKeys<"console">)
}

export function fixedSpecKeys(sku: ProductSKU | undefined, skus: ProductSKU[]) {
  const source = sku && skus.length > 1 ? [sku] : skus
  const keys = new Set<string>()
  for (const item of source) {
    for (const key of Object.keys(item.specs)) {
      keys.add(key)
    }
  }
  return [...keys].sort((left, right) => left.localeCompare(right))
}

export function initialSpecs(
  sku: ProductSKU | undefined,
  specsEditable: boolean,
  keys: string[],
) {
  if (specsEditable) {
    return sku
      ? Object.entries(sku.specs)
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([key, value]) => ({ key, value }))
      : [{ key: "", value: "" }]
  }
  return keys.map((key) => ({ key, value: sku?.specs[key] ?? "" }))
}

export function skuToInput(values: FormValues): ProductSKUInput {
  return {
    specs: Object.fromEntries(
      values.specs.map((spec) => [spec.key.trim(), spec.value.trim()]),
    ),
    price_cents: Math.round(Number(values.price.trim()) * 100),
    stock: Number(values.stock.trim()),
    is_enabled: values.is_enabled,
  }
}
