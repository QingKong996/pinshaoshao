import type { TFunction } from "i18next"
import { describe, expect, it } from "vitest"
import {
  MAX_SPECS,
  buildFormSchema,
  fixedSpecKeys,
  initialSpecs,
  skuToInput,
  translateValidationMessage,
} from "@/features/merchant-products/model/sku-form"
import type { ProductSKU } from "@/features/merchant-products/types"
import i18n from "@/shared/i18n/instance"

const sku: ProductSKU = {
  id: 10,
  product_id: 7,
  specs: { Size: "M", Color: "Blue" },
  price_cents: 1990,
  stock: 8,
  is_enabled: true,
  created_at: "2026-11-15T08:00:00Z",
  updated_at: "2026-11-15T09:00:00Z",
}

const validValues = {
  specs: [
    { key: "Color", value: "Blue" },
    { key: "Size", value: "M" },
  ],
  price: "19.90",
  stock: "8",
  is_enabled: true,
}

describe("SKU form schema", () => {
  it("accepts valid editable specs, price, and stock", () => {
    expect(buildFormSchema(true).safeParse(validValues).success).toBe(true)
  })

  it("rejects missing, duplicate, and empty specs", () => {
    const result = buildFormSchema(true).safeParse({
      ...validValues,
      specs: [
        { key: " ", value: "Blue" },
        { key: "Size", value: "" },
        { key: "Size", value: "M" },
      ],
    })
    const messages = result.error?.issues.map((issue) => issue.message)

    expect(messages).toEqual(
      expect.arrayContaining([
        "skuForm.errors.specKeyRequired",
        "skuForm.errors.specValueRequired",
        "skuForm.errors.specDuplicate",
      ]),
    )
  })

  it("rejects too many specs", () => {
    const result = buildFormSchema(true).safeParse({
      ...validValues,
      specs: Array.from({ length: MAX_SPECS + 1 }, (_, index) => ({
        key: `Spec ${index}`,
        value: "Value",
      })),
    })

    expect(result.error?.issues[0]).toMatchObject({
      message: "skuForm.errors.specLimit",
      path: ["specs", MAX_SPECS],
    })
  })

  it("rejects invalid prices and stock quantities", () => {
    expect(
      buildFormSchema(true).safeParse({
        ...validValues,
        price: "",
      }).error?.issues[0]?.message,
    ).toBe("skuForm.errors.priceRequired")
    expect(
      buildFormSchema(true).safeParse({
        ...validValues,
        price: "12.345",
      }).error?.issues[0]?.message,
    ).toBe("skuForm.errors.priceInvalid")
    expect(
      buildFormSchema(true).safeParse({
        ...validValues,
        stock: "",
      }).error?.issues[0]?.message,
    ).toBe("skuForm.errors.stockRequired")
    expect(
      buildFormSchema(true).safeParse({
        ...validValues,
        stock: "-1",
      }).error?.issues[0]?.message,
    ).toBe("skuForm.errors.stockInvalid")
  })

  it("keeps fixed spec keys read-only while still requiring values", () => {
    const result = buildFormSchema(false).safeParse({
      ...validValues,
      specs: [
        { key: "", value: "" },
        { key: "", value: "Blue" },
        { key: "", value: "Red" },
      ],
    })

    expect(result.error?.issues).toHaveLength(1)
    expect(result.error?.issues[0]?.message).toBe(
      "skuForm.errors.specValueRequired",
    )
  })
})

describe("SKU form helpers", () => {
  it("derives sorted fixed spec keys from all or one SKU", () => {
    const all = [sku, { ...sku, id: 11, specs: { Material: "Cotton" } }]

    expect(fixedSpecKeys(undefined, all)).toEqual(["Color", "Material", "Size"])
    expect(fixedSpecKeys(sku, all)).toEqual(["Color", "Size"])
  })

  it("builds initial editable and fixed specs", () => {
    expect(initialSpecs(sku, true, [])).toEqual([
      { key: "Color", value: "Blue" },
      { key: "Size", value: "M" },
    ])
    expect(initialSpecs(undefined, true, [])).toEqual([{ key: "", value: "" }])
    expect(initialSpecs(sku, false, ["Color", "Material"])).toEqual([
      { key: "Color", value: "Blue" },
      { key: "Material", value: "" },
    ])
  })

  it("converts form values to API input", () => {
    expect(
      skuToInput({
        specs: [
          { key: " Color ", value: " Blue " },
          { key: "Size", value: "M" },
        ],
        price: " 12.345 ",
        stock: " 12 ",
        is_enabled: false,
      }),
    ).toEqual({
      specs: { Color: "Blue", Size: "M" },
      price_cents: 1235,
      stock: 12,
      is_enabled: false,
    })
  })
})

describe("translateValidationMessage", () => {
  const t = i18n.getFixedT("en", "console") as TFunction<"console">

  it("translates the spec limit with interpolation", () => {
    expect(translateValidationMessage("skuForm.errors.specLimit", t)).toBe(
      "A SKU can have at most 5 specs.",
    )
    expect(translateValidationMessage(undefined, t)).toBeUndefined()
  })
})
