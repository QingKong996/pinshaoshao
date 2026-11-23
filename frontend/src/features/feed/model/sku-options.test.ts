import { describe, expect, it } from "vitest"
import type { ProductSKU } from "../types"
import {
  buildSKUDimensions,
  findSKUBySelection,
  isSKUOptionSelectable,
  updateSKUSelection,
} from "./sku-options"
import {
  findPreferredSKU,
  getSKUStatus,
  isSKUPurchasable,
} from "./product"

const skus: ProductSKU[] = [
  {
    id: 1,
    specs: { Color: "Black", Size: "M" },
    price_cents: 1999,
    is_enabled: true,
    in_stock: true,
  },
  {
    id: 2,
    specs: { Color: "Black", Size: "L" },
    price_cents: 2499,
    is_enabled: true,
    in_stock: false,
  },
  {
    id: 3,
    specs: { Color: "White", Size: "L" },
    price_cents: 2999,
    is_enabled: true,
    in_stock: true,
  },
  {
    id: 4,
    specs: { Color: "White", Size: "XL" },
    price_cents: 3499,
    is_enabled: false,
    in_stock: true,
  },
]

describe("buildSKUDimensions", () => {
  it("builds each specification dimension with aggregate option state", () => {
    const dimensions = buildSKUDimensions(skus, {})

    expect(dimensions?.map(({ name }) => name)).toEqual(["Color", "Size"])
    expect(
      dimensions?.find(({ name }) => name === "Color")?.options,
    ).toEqual([
      { value: "Black", status: "available" },
      { value: "White", status: "available" },
    ])
    expect(
      dimensions?.find(({ name }) => name === "Size")?.options,
    ).toEqual([
      { value: "M", status: "available" },
      { value: "L", status: "available" },
      { value: "XL", status: "unavailable" },
    ])
  })

  it("distinguishes out-of-stock, unavailable, and missing combinations", () => {
    const dimensions = buildSKUDimensions(skus, { Color: "Black" })
    const sizes = dimensions?.find(({ name }) => name === "Size")?.options

    expect(sizes?.find(({ value }) => value === "M")).toMatchObject({
      status: "available",
    })
    expect(sizes?.find(({ value }) => value === "L")).toMatchObject({
      status: "out_of_stock",
    })
    expect(sizes?.find(({ value }) => value === "XL")).toMatchObject({
      status: "missing",
    })

    const whiteDimensions = buildSKUDimensions(skus, { Color: "White" })
    expect(
      whiteDimensions
        ?.find(({ name }) => name === "Size")
        ?.options.find(({ value }) => value === "XL"),
    ).toMatchObject({ status: "unavailable" })
  })

  it("keeps existing out-of-stock options selectable", () => {
    const dimensions = buildSKUDimensions(skus, { Color: "Black" })
    const options = dimensions?.find(({ name }) => name === "Size")?.options

    expect(
      isSKUOptionSelectable(
        options!.find(({ value }) => value === "M")!,
      ),
    ).toBe(true)
    expect(
      isSKUOptionSelectable(
        options!.find(({ value }) => value === "L")!,
      ),
    ).toBe(true)
    expect(
      isSKUOptionSelectable(
        options!.find(({ value }) => value === "XL")!,
      ),
    ).toBe(false)
  })

  it("returns null when SKUs do not have consistent specifications", () => {
    expect(buildSKUDimensions([], {})).toBeNull()
    expect(
      buildSKUDimensions(
        [{ ...skus[0], specs: {} }],
        {},
      ),
    ).toBeNull()
    expect(
      buildSKUDimensions(
        [skus[0], { ...skus[1], specs: { Color: "Black" } }],
        {},
      ),
    ).toBeNull()
  })
})

describe("updateSKUSelection", () => {
  it("returns the current selection for an unknown dimension", () => {
    const selection = { Color: "Black" }

    expect(updateSKUSelection(skus, selection, "Material", "Linen")).toBe(
      selection,
    )
  })

  it("stops preserving later dimensions when no value was selected", () => {
    expect(
      updateSKUSelection(skus, { Color: "Black" }, "Color", "White"),
    ).toEqual({ Color: "White" })
  })

  it("preserves a later selection when it remains valid", () => {
    expect(
      updateSKUSelection(
        skus,
        { Color: "Black", Size: "L" },
        "Color",
        "White",
      ),
    ).toEqual({ Color: "White", Size: "L" })
  })

  it("clears later selections that no longer have a matching SKU", () => {
    expect(
      updateSKUSelection(
        skus,
        { Color: "Black", Size: "M" },
        "Color",
        "White",
      ),
    ).toEqual({ Color: "White" })
  })
})

describe("findSKUBySelection", () => {
  it("returns a SKU only for a complete matching selection", () => {
    expect(findSKUBySelection(skus, { Color: "White", Size: "L" })?.id).toBe(3)
    expect(findSKUBySelection(skus, { Color: "White" })).toBeNull()
    expect(
      findSKUBySelection(skus, { Color: "White", Size: "M" }),
    ).toBeNull()
  })
})

describe("SKU availability", () => {
  it("classifies purchasable, out-of-stock, and disabled SKUs", () => {
    expect(isSKUPurchasable(skus[0])).toBe(true)
    expect(isSKUPurchasable(skus[1])).toBe(false)
    expect(isSKUPurchasable(skus[3])).toBe(false)
    expect(getSKUStatus(skus[0])).toBe("available")
    expect(getSKUStatus(skus[1])).toBe("out_of_stock")
    expect(getSKUStatus(skus[3])).toBe("unavailable")
    expect(getSKUStatus({ ...skus[3], in_stock: false })).toBe("unavailable")
  })

  it("prefers an in-stock SKU, then an out-of-stock SKU, then a disabled one", () => {
    expect(findPreferredSKU(skus)?.id).toBe(1)
    expect(findPreferredSKU([skus[1], skus[3]])?.id).toBe(2)
    expect(findPreferredSKU([skus[3]])?.id).toBe(4)
    expect(findPreferredSKU([])).toBeNull()
  })
})
