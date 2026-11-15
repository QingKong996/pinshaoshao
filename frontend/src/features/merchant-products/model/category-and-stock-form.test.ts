import type { TFunction } from "i18next"
import { describe, expect, it } from "vitest"
import {
  categoryFormSchema,
  categoryToInput,
  categoryToValues,
  translateValidationMessage as translateCategoryMessage,
} from "@/features/merchant-products/model/category-form"
import {
  initialValues,
  stockFormSchema,
  translateValidationMessage as translateStockMessage,
} from "@/features/merchant-products/model/stock-form"
import i18n from "@/shared/i18n/instance"

const t = i18n.getFixedT("en", "console") as TFunction<"console">

describe("categoryFormSchema", () => {
  it("trims and validates category names", () => {
    expect(categoryFormSchema.parse({ name: "  Clothing  " })).toEqual({
      name: "Clothing",
    })
    expect(
      categoryFormSchema.safeParse({ name: " " }).error?.issues[0]?.message,
    ).toBe("categoryForm.errors.nameRequired")
    expect(
      categoryFormSchema.safeParse({ name: "x".repeat(51) }).error?.issues[0]
        ?.message,
    ).toBe("categoryForm.errors.nameLength")
  })

  it("converts category values to and from API input", () => {
    const category = {
      id: 1,
      name: "Clothing",
      product_count: 2,
      created_at: "2026-11-15T08:00:00Z",
      updated_at: "2026-11-15T09:00:00Z",
    }

    expect(categoryToValues(category)).toEqual({ name: "Clothing" })
    expect(categoryToInput({ name: "Clothing" })).toEqual({
      name: "Clothing",
    })
  })

  it("translates validation messages", () => {
    expect(
      translateCategoryMessage("categoryForm.errors.nameLength", t),
    ).toBe("The category name must be at most 50 characters.")
    expect(translateCategoryMessage(undefined, t)).toBeUndefined()
  })
})

describe("stockFormSchema", () => {
  it("accepts non-negative integers and builds initial values", () => {
    expect(stockFormSchema.parse({ stock: " 12 " })).toEqual({ stock: "12" })
    expect(stockFormSchema.parse({ stock: "0" })).toEqual({ stock: "0" })
    expect(initialValues(7)).toEqual({ stock: "7" })
  })

  it("rejects empty, non-integer, negative, and overflowing values", () => {
    const messages = [
      "",
      "1.5",
      "-1",
      "2147483648",
    ].map(
      (stock) =>
        stockFormSchema.safeParse({ stock }).error?.issues[0]?.message,
    )

    expect(messages).toEqual([
      "stockForm.errors.stockRequired",
      "stockForm.errors.stockInvalid",
      "stockForm.errors.stockInvalid",
      "stockForm.errors.stockInvalid",
    ])
  })

  it("translates stock validation messages", () => {
    expect(translateStockMessage("stockForm.errors.stockInvalid", t)).toBe(
      "Stock must be a valid non-negative whole number.",
    )
    expect(translateStockMessage(undefined, t)).toBeUndefined()
  })
})
