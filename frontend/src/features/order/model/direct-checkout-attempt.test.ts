import { webcrypto } from "node:crypto"
import { beforeEach, describe, expect, it } from "vitest"
import {
  checkoutSignature,
  directCheckoutSignature,
} from "./direct-checkout-attempt"

describe("checkout signatures", () => {
  beforeEach(() => {
    Object.defineProperty(globalThis.crypto, "subtle", {
      configurable: true,
      value: webcrypto.subtle,
    })
  })

  it("normalizes cart item order and delivery fields", async () => {
    const first = await checkoutSignature({
      items: [
        { cart_item_id: 9, quantity: 1 },
        { cart_item_id: 3, quantity: 2 },
      ],
      receiver_name: "  Alice ",
      receiver_phone: " 1234567890 ",
      shipping_address: " Shanghai ",
    })
    const second = await checkoutSignature({
      items: [
        { cart_item_id: 3, quantity: 2 },
        { cart_item_id: 9, quantity: 1 },
      ],
      receiver_name: "Alice",
      receiver_phone: "1234567890",
      shipping_address: "Shanghai",
    })

    expect(first).toBe(second)
  })

  it("normalizes direct purchase delivery fields", async () => {
    const first = await directCheckoutSignature({
      sku_id: 5,
      quantity: 2,
      receiver_name: " Alice ",
      receiver_phone: " 1234567890 ",
      shipping_address: " Shanghai ",
    })
    const second = await directCheckoutSignature({
      sku_id: 5,
      quantity: 2,
      receiver_name: "Alice",
      receiver_phone: "1234567890",
      shipping_address: "Shanghai",
    })

    expect(first).toBe(second)
  })
})
