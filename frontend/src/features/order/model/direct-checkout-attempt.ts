import type { CreateDirectOrderInput, CreateOrderInput } from "../types"

export type CheckoutAttempt = {
  signature: string
  key: string
}

export function cartCheckoutAttemptKey(username: string) {
  return `cart-checkout-attempt:${JSON.stringify(username)}`
}

export function directCheckoutAttemptKey(
  username: string,
  skuID: number,
  quantity: number,
) {
  return `direct-checkout-attempt:${JSON.stringify([username, skuID, quantity])}`
}

export async function directCheckoutSignature(
  payload: Omit<CreateDirectOrderInput, "checkout_key">,
) {
  return hashCheckoutPayload({
    sku_id: payload.sku_id,
    quantity: payload.quantity,
    ...normalizeCheckoutDelivery(payload),
  })
}

export async function checkoutSignature(
  payload: Omit<CreateOrderInput, "checkout_key">,
) {
  return hashCheckoutPayload({
    items: [...payload.items]
      .sort((left, right) => left.cart_item_id - right.cart_item_id)
      .map((item) => ({
        cart_item_id: item.cart_item_id,
        quantity: item.quantity,
      })),
    ...normalizeCheckoutDelivery(payload),
  })
}

function normalizeCheckoutDelivery(payload: {
  receiver_name: string
  receiver_phone: string
  shipping_address: string
}) {
  return {
    receiver_name: payload.receiver_name.trim(),
    receiver_phone: payload.receiver_phone.trim(),
    shipping_address: payload.shipping_address.trim(),
  }
}

async function hashCheckoutPayload(payload: unknown) {
  const bytes = new TextEncoder().encode(JSON.stringify(payload))
  const digest = await crypto.subtle.digest("SHA-256", bytes)
  return Array.from(new Uint8Array(digest), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("")
}

export function readCheckoutAttempt(
  storageKey: string,
): CheckoutAttempt | null {
  try {
    const raw = sessionStorage.getItem(storageKey)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<CheckoutAttempt> | null
    if (
      typeof parsed?.signature !== "string" ||
      !/^[a-f0-9]{64}$/.test(parsed.signature) ||
      typeof parsed.key !== "string" ||
      !/^[a-f0-9]{8}(?:-[a-f0-9]{4}){3}-[a-f0-9]{12}$/i.test(parsed.key)
    ) {
      return null
    }
    return { signature: parsed.signature, key: parsed.key }
  } catch {
    return null
  }
}

export function saveCheckoutAttempt(
  storageKey: string,
  attempt: CheckoutAttempt | null,
) {
  try {
    if (attempt) {
      sessionStorage.setItem(storageKey, JSON.stringify(attempt))
    } else {
      sessionStorage.removeItem(storageKey)
    }
  } catch {
    return
  }
}

export function readDirectCheckoutAttempt(storageKey: string) {
  return readCheckoutAttempt(storageKey)
}

export function saveDirectCheckoutAttempt(
  storageKey: string,
  attempt: CheckoutAttempt | null,
) {
  saveCheckoutAttempt(storageKey, attempt)
}
