import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"

const USERNAME_MIN = 5
const USERNAME_MAX = 20
const PASSWORD_MIN = 8
const PASSWORD_MAX = 20

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const VALIDATION_KEYS = {
  usernameRequired: "auth:validation.usernameRequired",
  usernameLength: "auth:validation.usernameLength",
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

const usernameSchema = z
  .string()
  .trim()
  .min(1, VALIDATION_KEYS.usernameRequired)
  .min(USERNAME_MIN, VALIDATION_KEYS.usernameLength)
  .max(USERNAME_MAX, VALIDATION_KEYS.usernameLength)

const passwordSchema = z
  .string()
  .min(1, VALIDATION_KEYS.passwordRequired)
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
  .min(1, VALIDATION_KEYS.emailRequired)
  .refine((value) => EMAIL_PATTERN.test(value), VALIDATION_KEYS.emailInvalid)

function createSignUpSchema(requireEmail: boolean) {
  return z
    .object({
      username: usernameSchema,
      email: requireEmail ? requiredEmailSchema : optionalEmailSchema,
      password: passwordSchema,
      confirmPassword: z
        .string()
        .min(1, VALIDATION_KEYS.confirmPasswordRequired),
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

export const merchantSignUpSchema = createSignUpSchema(true)
