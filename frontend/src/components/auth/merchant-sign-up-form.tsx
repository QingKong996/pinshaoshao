import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router"
import { AuthErrorAlert } from "@/components/auth/auth-error-alert"
import { PasswordInput } from "@/components/auth/password-input"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  merchantSignUpSchema,
  translateAuthValidationMessage,
  type MerchantSignUpValues,
} from "@/auth/schemas"
import { describeMerchantAuthError } from "@/auth/merchant/errors"
import { useMerchantAuth } from "@/auth/merchant/context"
import { MERCHANT_CONSOLE_HOME_PATH } from "@/merchant-console/nav"

const INITIAL_VALUES: MerchantSignUpValues = {
  name: "",
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
}

export function MerchantSignUpForm() {
  const { t } = useTranslation()
  const { session, signUp } = useMerchantAuth()
  const [serverError, setServerError] = useState<unknown>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<MerchantSignUpValues>({
    resolver: zodResolver(merchantSignUpSchema),
    defaultValues: INITIAL_VALUES,
  })

  async function onSubmit(values: MerchantSignUpValues) {
    setServerError(null)
    try {
      await signUp(values)
    } catch (error) {
      setServerError(error)
    }
  }

  function clearServerError() {
    setServerError(null)
  }

  if (session) {
    return <Navigate to={MERCHANT_CONSOLE_HOME_PATH} replace />
  }

  return (
    <form
      noValidate
      aria-busy={isSubmitting}
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col gap-5"
    >
      {serverError ? (
        <AuthErrorAlert
          message={describeMerchantAuthError(serverError, "register", t)}
        />
      ) : null}

      <FieldGroup>
        <Field data-invalid={errors.name ? true : undefined}>
          <FieldLabel htmlFor="merchant-signup-shop-name">
            {t("auth:merchant.fields.name")}
          </FieldLabel>
          <Input
            id="merchant-signup-shop-name"
            autoComplete="organization"
            autoFocus
            required
            aria-invalid={errors.name ? true : undefined}
            aria-describedby={
              errors.name
                ? "merchant-signup-shop-name-hint merchant-signup-shop-name-error"
                : "merchant-signup-shop-name-hint"
            }
            className="h-10"
            {...register("name", { onChange: clearServerError })}
          />
          <FieldDescription id="merchant-signup-shop-name-hint">
            {t("auth:merchant.hints.name")}
          </FieldDescription>
          {errors.name ? (
            <FieldError id="merchant-signup-shop-name-error">
              {translateAuthValidationMessage(errors.name.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.username ? true : undefined}>
          <FieldLabel htmlFor="merchant-signup-username">
            {t("auth:merchant.fields.username")}
          </FieldLabel>
          <Input
            id="merchant-signup-username"
            autoComplete="username"
            autoCapitalize="none"
            spellCheck={false}
            required
            aria-invalid={errors.username ? true : undefined}
            aria-describedby={
              errors.username
                ? "merchant-signup-username-hint merchant-signup-username-error"
                : "merchant-signup-username-hint"
            }
            className="h-10"
            {...register("username", { onChange: clearServerError })}
          />
          <FieldDescription id="merchant-signup-username-hint">
            {t("auth:hints.username")}
          </FieldDescription>
          {errors.username ? (
            <FieldError id="merchant-signup-username-error">
              {translateAuthValidationMessage(errors.username.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.email ? true : undefined}>
          <FieldLabel htmlFor="merchant-signup-email">
            {t("auth:fields.email")}
          </FieldLabel>
          <Input
            id="merchant-signup-email"
            type="email"
            autoComplete="email"
            autoCapitalize="none"
            spellCheck={false}
            required
            aria-invalid={errors.email ? true : undefined}
            aria-describedby={
              errors.email
                ? "merchant-signup-email-hint merchant-signup-email-error"
                : "merchant-signup-email-hint"
            }
            className="h-10"
            {...register("email", { onChange: clearServerError })}
          />
          <FieldDescription id="merchant-signup-email-hint">
            {t("auth:merchant.hints.email")}
          </FieldDescription>
          {errors.email ? (
            <FieldError id="merchant-signup-email-error">
              {translateAuthValidationMessage(errors.email.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.password ? true : undefined}>
          <FieldLabel htmlFor="merchant-signup-password">
            {t("auth:fields.password")}
          </FieldLabel>
          <PasswordInput
            id="merchant-signup-password"
            autoComplete="new-password"
            required
            aria-invalid={errors.password ? true : undefined}
            aria-describedby={
              errors.password
                ? "merchant-signup-password-hint merchant-signup-password-error"
                : "merchant-signup-password-hint"
            }
            className="h-10"
            {...register("password", {
              deps: ["confirmPassword"],
              onChange: clearServerError,
            })}
          />
          <FieldDescription id="merchant-signup-password-hint">
            {t("auth:hints.password")}
          </FieldDescription>
          {errors.password ? (
            <FieldError id="merchant-signup-password-error">
              {translateAuthValidationMessage(errors.password.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.confirmPassword ? true : undefined}>
          <FieldLabel htmlFor="merchant-signup-confirm-password">
            {t("auth:fields.confirmPassword")}
          </FieldLabel>
          <PasswordInput
            id="merchant-signup-confirm-password"
            autoComplete="new-password"
            required
            aria-invalid={errors.confirmPassword ? true : undefined}
            aria-describedby={
              errors.confirmPassword
                ? "merchant-signup-confirm-password-error"
                : undefined
            }
            className="h-10"
            {...register("confirmPassword", { onChange: clearServerError })}
          />
          {errors.confirmPassword ? (
            <FieldError id="merchant-signup-confirm-password-error">
              {translateAuthValidationMessage(
                errors.confirmPassword.message,
                t,
              )}
            </FieldError>
          ) : null}
        </Field>
      </FieldGroup>

      <Button
        type="submit"
        size="lg"
        disabled={isSubmitting}
        className="h-10 w-full gap-2"
      >
        {isSubmitting ? (
          <>
            <LoaderCircle className="animate-spin" aria-hidden />
            {t("auth:merchant.signUp.submitting")}
          </>
        ) : (
          t("auth:merchant.signUp.submit")
        )}
      </Button>

      <span role="status" className="sr-only">
        {isSubmitting ? t("auth:merchant.signUp.status") : ""}
      </span>
    </form>
  )
}
