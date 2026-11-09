import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"

const REQUIRED_MIN = 1
const USERNAME_MIN = 5
const USERNAME_MAX = 20
const PASSWORD_MIN = 8
const PASSWORD_MAX = 20
const SHOP_NAME_MIN = 2
const SHOP_NAME_MAX = 50

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const VALIDATION_KEYS = {
  usernameRequired: "auth:validation.usernameRequired",
  usernameLength: "auth:validation.usernameLength",
  shopNameRequired: "auth:validation.shopNameRequired",
  shopNameLength: "auth:validation.shopNameLength",
  emailRequired: "auth:validation.emailRequired",
  emailInvalid: "auth:validation.emailInvalid",
  passwordRequired: "auth:validation.passwordRequired",
  passwordLength: "auth:validation.passwordLength",
  confirmPasswordRequired: "auth:validation.confirmPasswordRequired",
  confirmPasswordMismatch: "auth:validation.confirmPasswordMismatch",
} as const satisfies Record<string, ParseKeys>

const LENGTH_MESSAGE_VALUES: Partial<
  Record<ParseKeys, { min: number; max: number }>
> = {
  [VALIDATION_KEYS.usernameLength]: { min: USERNAME_MIN, max: USERNAME_MAX },
  [VALIDATION_KEYS.passwordLength]: { min: PASSWORD_MIN, max: PASSWORD_MAX },
  [VALIDATION_KEYS.shopNameLength]: {
    min: SHOP_NAME_MIN,
    max: SHOP_NAME_MAX,
  },
}

export function translateAuthValidationMessage(
  message: string | undefined,
  t: TFunction,
): string | undefined {
  if (!message) {
    return undefined
  }
  const key = message as ParseKeys
  const values = LENGTH_MESSAGE_VALUES[key]
  return values ? t(key, values) : t(key)
}

type BoundedStringKeys = {
  required: ParseKeys
  length: ParseKeys
}

function requiredBoundedString(
  min: number,
  max: number,
  keys: BoundedStringKeys,
) {
  return z
    .string()
    .trim()
    .min(REQUIRED_MIN, keys.required)
    .min(min, keys.length)
    .max(max, keys.length)
}

const usernameSchema = requiredBoundedString(USERNAME_MIN, USERNAME_MAX, {
  required: VALIDATION_KEYS.usernameRequired,
  length: VALIDATION_KEYS.usernameLength,
})

const shopNameSchema = requiredBoundedString(SHOP_NAME_MIN, SHOP_NAME_MAX, {
  required: VALIDATION_KEYS.shopNameRequired,
  length: VALIDATION_KEYS.shopNameLength,
})

const passwordSchema = z
  .string()
  .min(REQUIRED_MIN, VALIDATION_KEYS.passwordRequired)
  .min(PASSWORD_MIN, VALIDATION_KEYS.passwordLength)
  .max(PASSWORD_MAX, VALIDATION_KEYS.passwordLength)

const optionalEmailSchema = z
  .string()
  .trim()
  .refine(
    (value) => !value || EMAIL_PATTERN.test(value),
    VALIDATION_KEYS.emailInvalid,
  )

const requiredEmailSchema = z
  .string()
  .trim()
  .min(REQUIRED_MIN, VALIDATION_KEYS.emailRequired)
  .refine((value) => EMAIL_PATTERN.test(value), VALIDATION_KEYS.emailInvalid)

function createSignUpSchema(requireEmail: boolean) {
  return z
    .object({
      username: usernameSchema,
      email: requireEmail ? requiredEmailSchema : optionalEmailSchema,
      password: passwordSchema,
      confirmPassword: z
        .string()
        .min(REQUIRED_MIN, VALIDATION_KEYS.confirmPasswordRequired),
    })
    .refine((values) => values.confirmPassword === values.password, {
      error: VALIDATION_KEYS.confirmPasswordMismatch,
      path: ["confirmPassword"],
    })
}

export const signInSchema = z.object({
  username: usernameSchema,
  password: passwordSchema,
})

export type SignInValues = z.infer<typeof signInSchema>

export const signUpSchema = createSignUpSchema(false)

export type SignUpValues = z.infer<typeof signUpSchema>

export const merchantSignUpSchema = z
  .object({
    name: shopNameSchema,
    username: usernameSchema,
    email: requiredEmailSchema,
    password: passwordSchema,
    confirmPassword: z
      .string()
      .min(REQUIRED_MIN, VALIDATION_KEYS.confirmPasswordRequired),
  })
  .refine((values) => values.confirmPassword === values.password, {
    error: VALIDATION_KEYS.confirmPasswordMismatch,
    path: ["confirmPassword"],
  })

export type MerchantSignUpValues = z.infer<typeof merchantSignUpSchema>
