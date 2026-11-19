import type { TFunction } from "i18next"
import { describe, expect, it } from "vitest"
import {
  INITIAL_VALUES,
  productFormSchema,
  productToInput,
  productToValues,
  translateValidationMessage,
  UNCATEGORIZED,
} from "@/features/merchant-products/model/product-form"
import type { Product } from "@/features/merchant-products/types"
import i18n from "@/shared/i18n/instance"

const product: Product = {
  id: 7,
  name: "Cotton T-shirt",
  description: "Soft cotton",
  status: "on_shelf",
  category_id: 3,
  category_name: "Clothing",
  sku_count: 2,
  min_price_cents: 1990,
  max_price_cents: 2590,
  created_at: "2026-11-15T08:00:00Z",
  updated_at: "2026-11-15T09:00:00Z",
}

describe("productFormSchema", () => {
  it("trims names and accepts valid product input", () => {
    const result = productFormSchema.parse({
      ...INITIAL_VALUES,
      name: "  Cotton T-shirt  ",
    })

    expect(result.name).toBe("Cotton T-shirt")
  })

  it("rejects blank and oversized names", () => {
    expect(
      productFormSchema.safeParse({ ...INITIAL_VALUES, name: "   " }).error
        ?.issues[0]?.message,
    ).toBe("productForm.errors.nameRequired")
    expect(
      productFormSchema.safeParse({
        ...INITIAL_VALUES,
        name: "x".repeat(101),
      }).error?.issues[0]?.message,
    ).toBe("productForm.errors.nameLength")
  })
})

describe("product form conversions", () => {
  it("maps a product into editable values", () => {
    expect(productToValues(product)).toEqual({
      name: "Cotton T-shirt",
      description: "Soft cotton",
      categoryId: "3",
      status: "on_shelf",
    })
  })

  it("maps uncategorized products and form values back to API input", () => {
    expect(
      productToValues({
        ...product,
        category_id: null,
      }).categoryId,
    ).toBe(UNCATEGORIZED)
    expect(
      productToInput({
        name: "Cotton T-shirt",
        description: "Soft cotton",
        categoryId: "3",
        status: "off_shelf",
      }),
    ).toEqual({
      name: "Cotton T-shirt",
      description: "Soft cotton",
      category_id: 3,
      status: "off_shelf",
    })
    expect(
      productToInput({
        ...INITIAL_VALUES,
        name: "Uncategorized",
      }).category_id,
    ).toBeNull()
  })
})

describe("translateValidationMessage", () => {
  const t = i18n.getFixedT("en", "console") as TFunction<"console">

  it("translates length and simple validation messages", () => {
    expect(
      translateValidationMessage("productForm.errors.nameLength", t),
    ).toBe("The product name must be at most 100 characters.")
    expect(
      translateValidationMessage("productForm.errors.nameRequired", t),
    ).toBe("Enter a product name.")
    expect(translateValidationMessage(undefined, t)).toBeUndefined()
  })
})
