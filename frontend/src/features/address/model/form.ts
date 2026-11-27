import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"
import type { Address, AddressInput } from "../types"

const VALIDATION_MESSAGES = {
  receiverNameRequired: "errors.receiverNameRequired",
  receiverNameLength: "errors.receiverNameLength",
  receiverPhoneRequired: "errors.receiverPhoneRequired",
  receiverPhoneLength: "errors.receiverPhoneLength",
  shippingAddressRequired: "errors.shippingAddressRequired",
  shippingAddressLength: "errors.shippingAddressLength",
} as const

export const addressFormSchema = z.object({
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
  is_default: z.boolean(),
})

export type AddressFormValues = z.infer<typeof addressFormSchema>

export const INITIAL_ADDRESS_VALUES: AddressFormValues = {
  receiver_name: "",
  receiver_phone: "",
  shipping_address: "",
  is_default: false,
}

export function addressToFormValues(address: Address): AddressFormValues {
  return {
    receiver_name: address.receiver_name,
    receiver_phone: address.receiver_phone,
    shipping_address: address.shipping_address,
    is_default: address.is_default,
  }
}

export function addressToInput(values: AddressFormValues): AddressInput {
  return {
    receiver_name: values.receiver_name,
    receiver_phone: values.receiver_phone,
    shipping_address: values.shipping_address,
    is_default: values.is_default,
  }
}

export function translateAddressValidationMessage(
  message: string | undefined,
  t: TFunction<"address">,
) {
  if (!message) {
    return undefined
  }
  return t(message as ParseKeys<"address">)
}
