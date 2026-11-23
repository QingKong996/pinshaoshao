import type { TFunction } from "i18next"
import { describe, expect, it, vi } from "vitest"
import {
  checkoutFormSchema,
  translateCheckoutValidationMessage,
} from "./checkout-form"

const validValues = {
  receiver_name: "Receiver",
  receiver_phone: "1234567890",
  shipping_address: "Shanghai",
}

describe("checkoutFormSchema", () => {
  it("trims valid delivery details", () => {
    expect(
      checkoutFormSchema.parse({
        receiver_name: "  Receiver  ",
        receiver_phone: " 1234567890 ",
        shipping_address: " Shanghai ",
      }),
    ).toEqual(validValues)
  })

  it.each([
    {
      field: "receiver_name" as const,
      value: " ",
      message: "checkout.errors.receiverNameRequired",
    },
    {
      field: "receiver_name" as const,
      value: "a".repeat(51),
      message: "checkout.errors.receiverNameLength",
    },
    {
      field: "receiver_phone" as const,
      value: "",
      message: "checkout.errors.receiverPhoneRequired",
    },
    {
      field: "receiver_phone" as const,
      value: "1".repeat(33),
      message: "checkout.errors.receiverPhoneLength",
    },
    {
      field: "shipping_address" as const,
      value: " ",
      message: "checkout.errors.shippingAddressRequired",
    },
    {
      field: "shipping_address" as const,
      value: "a".repeat(1001),
      message: "checkout.errors.shippingAddressLength",
    },
  ])("rejects invalid $field", ({ field, value, message }) => {
    const result = checkoutFormSchema.safeParse({
      ...validValues,
      [field]: value,
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(message)
    }
  })
})

describe("translateCheckoutValidationMessage", () => {
  it("translates known validation keys and ignores missing messages", () => {
    const t = vi.fn((key: string) => `translated:${key}`) as unknown as TFunction<
      "order"
    >

    expect(translateCheckoutValidationMessage(undefined, t)).toBeUndefined()
    expect(
      translateCheckoutValidationMessage(
        "checkout.errors.receiverNameRequired",
        t,
      ),
    ).toBe("translated:checkout.errors.receiverNameRequired")
    expect(t).toHaveBeenCalledOnce()
  })
})
