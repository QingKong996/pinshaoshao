import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"
import type { CreateMockShipmentInput } from "../dto/shipment"

const VALIDATION_MESSAGES = {
  merchantOrderNoRequired: "validation.merchantOrderNoRequired",
  merchantOrderNoLength: "validation.merchantOrderNoLength",
  notifyRequired: "validation.notifyRequired",
  notifyLength: "validation.notifyLength",
  notifyInvalid: "validation.notifyInvalid",
  itemSummaryLength: "validation.itemSummaryLength",
  carrierLength: "validation.carrierLength",
  cityLength: "validation.cityLength",
  receiverNameLength: "validation.receiverNameLength",
  receiverPhoneLength: "validation.receiverPhoneLength",
  receiverAddressLength: "validation.receiverAddressLength",
} as const

function isHttpUrl(value: string) {
  try {
    const url = new URL(value)
    return (
      (url.protocol === "http:" || url.protocol === "https:") &&
      url.hostname.length > 0 &&
      url.username.length === 0 &&
      url.password.length === 0 &&
      url.hash.length === 0
    )
  } catch {
    return false
  }
}

export const mockShipmentFormSchema = z.object({
  merchant_order_no: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.merchantOrderNoRequired)
    .max(64, VALIDATION_MESSAGES.merchantOrderNoLength),
  item_summary: z
    .string()
    .trim()
    .max(200, VALIDATION_MESSAGES.itemSummaryLength),
  carrier: z.string().trim().max(64, VALIDATION_MESSAGES.carrierLength),
  origin_city: z.string().trim().max(64, VALIDATION_MESSAGES.cityLength),
  destination_city: z
    .string()
    .trim()
    .max(64, VALIDATION_MESSAGES.cityLength),
  receiver_name: z
    .string()
    .trim()
    .max(100, VALIDATION_MESSAGES.receiverNameLength),
  receiver_phone: z
    .string()
    .trim()
    .max(32, VALIDATION_MESSAGES.receiverPhoneLength),
  receiver_address: z
    .string()
    .trim()
    .max(500, VALIDATION_MESSAGES.receiverAddressLength),
  notify_url: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.notifyRequired)
    .max(2048, VALIDATION_MESSAGES.notifyLength)
    .refine(isHttpUrl, VALIDATION_MESSAGES.notifyInvalid),
})

export type MockShipmentFormValues = z.infer<typeof mockShipmentFormSchema>

export function createMockShipmentInitialValues(): MockShipmentFormValues {
  return {
    merchant_order_no: `sandbox-${Date.now()}`,
    item_summary: "",
    carrier: "",
    origin_city: "",
    destination_city: "",
    receiver_name: "",
    receiver_phone: "",
    receiver_address: "",
    notify_url: "http://localhost:8080/mocklogistics/webhook",
  }
}

export function toCreateMockShipmentInput(
  values: MockShipmentFormValues,
): CreateMockShipmentInput {
  return {
    merchant_order_no: values.merchant_order_no.trim(),
    notify_url: values.notify_url.trim(),
    ...optionalValue("item_summary", values.item_summary),
    ...optionalValue("carrier", values.carrier),
    ...optionalValue("origin_city", values.origin_city),
    ...optionalValue("destination_city", values.destination_city),
    ...optionalValue("receiver_name", values.receiver_name),
    ...optionalValue("receiver_phone", values.receiver_phone),
    ...optionalValue("receiver_address", values.receiver_address),
  }
}

function optionalValue<Key extends keyof CreateMockShipmentInput>(
  key: Key,
  value: string,
) {
  const normalized = value.trim()
  return normalized ? { [key]: normalized } : {}
}

export function translateMockShipmentValidationMessage(
  message: string | undefined,
  t: TFunction<"mocklogistics">,
) {
  if (!message) {
    return undefined
  }
  return t(message as ParseKeys<"mocklogistics">)
}
