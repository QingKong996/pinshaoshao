import type { ProductSKU } from "../types"
import {
  getSKUStatus,
  isSKUPurchasable,
  type ProductSKUStatus,
} from "./product"

export type ProductSKUSelection = Record<string, string>

export type ProductSKUOption = {
  value: string
  status: ProductSKUStatus | "missing"
}

export type ProductSKUDimension = {
  name: string
  options: ProductSKUOption[]
}

export function buildSKUDimensions(
  skus: ProductSKU[],
  selection: ProductSKUSelection,
): ProductSKUDimension[] | null {
  const names = getSpecNames(skus)
  if (!names) {
    return null
  }

  return names.map((name, index) => {
    const values = Array.from(
      new Set(skus.map((sku) => sku.specs[name])),
    )
    return {
      name,
      options: values.map((value) => {
        const candidates = skus.filter(
          (sku) =>
            matchesPriorSpecs(sku, names, selection, index) &&
            sku.specs[name] === value,
        )
        return {
          value,
          status: getOptionStatus(candidates),
        }
      }),
    }
  })
}

export function updateSKUSelection(
  skus: ProductSKU[],
  selection: ProductSKUSelection,
  dimensionName: string,
  value: string,
): ProductSKUSelection {
  const names = getSpecNames(skus)
  const dimensionIndex = names?.indexOf(dimensionName) ?? -1
  if (!names || dimensionIndex < 0) {
    return selection
  }

  const next: ProductSKUSelection = {}
  for (const name of names.slice(0, dimensionIndex)) {
    const previousValue = selection[name]
    if (previousValue !== undefined) {
      next[name] = previousValue
    }
  }
  next[dimensionName] = value

  for (let index = dimensionIndex + 1; index < names.length; index += 1) {
    const name = names[index]
    const previousValue = selection[name]
    if (previousValue === undefined) {
      break
    }
    if (
      !skus.some(
        (sku) =>
          matchesPriorSpecs(sku, names, next, index) &&
          sku.specs[name] === previousValue,
      )
    ) {
      break
    }
    next[name] = previousValue
  }

  return next
}

export function findSKUBySelection(
  skus: ProductSKU[],
  selection: ProductSKUSelection,
) {
  const names = getSpecNames(skus)
  if (!names || names.some((name) => selection[name] === undefined)) {
    return null
  }
  return (
    skus.find((sku) =>
      names.every((name) => sku.specs[name] === selection[name]),
    ) ?? null
  )
}

export function isSKUOptionSelectable(option: ProductSKUOption) {
  return option.status !== "missing" && option.status !== "unavailable"
}

function getSpecNames(skus: ProductSKU[]) {
  if (skus.length === 0) {
    return null
  }
  const names = Object.keys(skus[0].specs).sort((left, right) =>
    left.localeCompare(right),
  )
  if (names.length === 0) {
    return null
  }

  const isConsistent = skus.every((sku) => {
    const keys = Object.keys(sku.specs)
    return (
      keys.length === names.length &&
      names.every(
        (name) =>
          typeof sku.specs[name] === "string" && sku.specs[name].length > 0,
      )
    )
  })
  return isConsistent ? names : null
}

function matchesPriorSpecs(
  sku: ProductSKU,
  names: string[],
  selection: ProductSKUSelection,
  endExclusive: number,
) {
  return names
    .slice(0, endExclusive)
    .every(
      (name) =>
        selection[name] === undefined || sku.specs[name] === selection[name],
    )
}

function getOptionStatus(
  candidates: ProductSKU[],
): ProductSKUStatus | "missing" {
  if (candidates.length === 0) {
    return "missing"
  }
  if (candidates.some(isSKUPurchasable)) {
    return "available"
  }
  if (candidates.some((sku) => getSKUStatus(sku) === "out_of_stock")) {
    return "out_of_stock"
  }
  return "unavailable"
}
