import type { TFunction } from "i18next"
import { describe, expect, it } from "vitest"
import {
  merchantSignUpSchema,
  signInSchema,
  signUpSchema,
  translateAuthValidationMessage,
} from "@/features/auth/model/schemas"
import i18n from "@/shared/i18n/instance"

const t = i18n.getFixedT("en", "auth") as TFunction
const validPassword = "a".repeat(12)

describe("signInSchema", () => {
  it("accepts a valid login", () => {
    expect(
      signInSchema.safeParse({
        username: "shopper",
        password: validPassword,
      }).success,
    ).toBe(true)
  })

  it("rejects blank, short, and oversized credentials", () => {
    const result = signInSchema.safeParse({
      username: "   ",
      password: "short",
    })

    expect(result.success).toBe(false)
    expect(result.error?.issues.map((issue) => issue.message)).toEqual([
      "auth:validation.usernameRequired",
      "auth:validation.usernameLength",
      "auth:validation.passwordLength",
    ])
  })
})

describe("signUpSchema", () => {
  it("allows an omitted email", () => {
    expect(
      signUpSchema.safeParse({
        username: "shopper",
        email: "",
        password: validPassword,
        confirmPassword: validPassword,
      }).success,
    ).toBe(true)
  })

  it("rejects an invalid email", () => {
    const result = signUpSchema.safeParse({
      username: "shopper",
      email: "not-an-email",
      password: validPassword,
      confirmPassword: validPassword,
    })

    expect(result.error?.issues[0]?.message).toBe(
      "auth:validation.emailInvalid",
    )
  })

  it("rejects mismatched passwords on the confirmation field", () => {
    const result = signUpSchema.safeParse({
      username: "shopper",
      email: "",
      password: validPassword,
      confirmPassword: "b".repeat(12),
    })

    expect(result.error?.issues).toContainEqual(
      expect.objectContaining({
        message: "auth:validation.confirmPasswordMismatch",
        path: ["confirmPassword"],
      }),
    )
  })
})

describe("merchantSignUpSchema", () => {
  it("requires a shop name and email", () => {
    const result = merchantSignUpSchema.safeParse({
      name: "",
      username: "merchant",
      email: "",
      password: validPassword,
      confirmPassword: validPassword,
    })

    expect(result.error?.issues.map((issue) => issue.message)).toEqual(
      expect.arrayContaining([
        "auth:validation.shopNameRequired",
        "auth:validation.emailRequired",
      ]),
    )
  })

  it("accepts valid merchant registration data", () => {
    expect(
      merchantSignUpSchema.safeParse({
        name: "Example Shop",
        username: "merchant",
        email: "owner@example.com",
        password: validPassword,
        confirmPassword: validPassword,
      }).success,
    ).toBe(true)
  })
})

describe("translateAuthValidationMessage", () => {
  it("interpolates the configured length range", () => {
    expect(
      translateAuthValidationMessage(
        "auth:validation.usernameLength",
        t,
      ),
    ).toBe("Username must be 5-20 characters")
  })

  it("returns undefined for a missing message", () => {
    expect(translateAuthValidationMessage(undefined, t)).toBeUndefined()
  })
})
