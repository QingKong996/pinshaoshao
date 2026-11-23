import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"

const VALIDATION_MESSAGES = {
  receiverNameRequired: "checkout.errors.receiverNameRequired",
  receiverNameLength: "checkout.errors.receiverNameLength",
  receiverPhoneRequired: "checkout.errors.receiverPhoneRequired",
  receiverPhoneLength: "checkout.errors.receiverPhoneLength",
  shippingAddressRequired: "checkout.errors.shippingAddressRequired",
  shippingAddressLength: "checkout.errors.shippingAddressLength",
} as const

export const checkoutFormSchema = z.object({
  receiver_name: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.receiverNameRequired)
    .max(50, VALIDATION_MESSAGES.receiverNameLength),
  receiver_phone: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.receiverPhoneRequired)
    .max(32, VALIDATION_MESSAGES.receiverPhoneLength),
  shipping_address: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.shippingAddressRequired)
    .max(1000, VALIDATION_MESSAGES.shippingAddressLength),
})

export type CheckoutFormValues = z.infer<typeof checkoutFormSchema>

export const INITIAL_CHECKOUT_VALUES: CheckoutFormValues = {
  receiver_name: "",
  receiver_phone: "",
  shipping_address: "",
}

export function translateCheckoutValidationMessage(
  message: string | undefined,
  t: TFunction<"order">,
) {
  if (!message) {
    return undefined
  }
  return t(message as ParseKeys<"order">)
}
