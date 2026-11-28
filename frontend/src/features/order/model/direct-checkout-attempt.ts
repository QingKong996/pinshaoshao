import type { CreateDirectOrderInput } from "../types"

type DirectCheckoutAttempt = {
  signature: string
  key: string
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
  const bytes = new TextEncoder().encode(JSON.stringify(payload))
  const digest = await crypto.subtle.digest("SHA-256", bytes)
  return Array.from(new Uint8Array(digest), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("")
}

export function readDirectCheckoutAttempt(
  storageKey: string,
): DirectCheckoutAttempt | null {
  try {
    const raw = sessionStorage.getItem(storageKey)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<DirectCheckoutAttempt> | null
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

export function saveDirectCheckoutAttempt(
  storageKey: string,
  attempt: DirectCheckoutAttempt | null,
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
