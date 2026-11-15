import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router"
import { AuthErrorAlert } from "./auth-error-alert"
import { AuthSubmitButton, AuthTextField } from "./auth-form-controls"
import { FieldGroup } from "@/shared/ui/field"
import {
  merchantSignUpSchema,
  translateAuthValidationMessage,
  type MerchantSignUpValues,
} from "../model/schemas"
import { describeMerchantAuthError } from "../model/merchant-errors"
import { useMerchantAuth } from "../model/merchant-context"

const INITIAL_VALUES: MerchantSignUpValues = {
  name: "",
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
}

export function MerchantSignUpForm({ redirectTo }: { redirectTo: string }) {
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
    return <Navigate to={redirectTo} replace />
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
        <AuthTextField
          id="merchant-signup-shop-name"
          label={t("auth:merchant.fields.name")}
          autoComplete="organization"
          autoFocus
          hint={t("auth:merchant.hints.name")}
          error={translateAuthValidationMessage(errors.name?.message, t)}
          registerProps={register("name", {
            onChange: clearServerError,
          })}
        />
        <AuthTextField
          id="merchant-signup-username"
          label={t("auth:merchant.fields.username")}
          autoComplete="username"
          autoCapitalize="none"
          spellCheck={false}
          hint={t("auth:hints.username")}
          error={translateAuthValidationMessage(
            errors.username?.message,
            t,
          )}
          registerProps={register("username", {
            onChange: clearServerError,
          })}
        />
        <AuthTextField
          id="merchant-signup-email"
          label={t("auth:fields.email")}
          type="email"
          autoComplete="email"
          autoCapitalize="none"
          spellCheck={false}
          hint={t("auth:merchant.hints.email")}
          error={translateAuthValidationMessage(
            errors.email?.message,
            t,
          )}
          registerProps={register("email", {
            onChange: clearServerError,
          })}
        />
        <AuthTextField
          id="merchant-signup-password"
          label={t("auth:fields.password")}
          type="password"
          autoComplete="new-password"
          hint={t("auth:hints.password")}
          error={translateAuthValidationMessage(
            errors.password?.message,
            t,
          )}
          registerProps={register("password", {
            deps: ["confirmPassword"],
            onChange: clearServerError,
          })}
        />
        <AuthTextField
          id="merchant-signup-confirm-password"
          label={t("auth:fields.confirmPassword")}
          type="password"
          autoComplete="new-password"
          error={translateAuthValidationMessage(
            errors.confirmPassword?.message,
            t,
          )}
          registerProps={register("confirmPassword", {
            onChange: clearServerError,
          })}
        />
      </FieldGroup>

      <AuthSubmitButton
        isSubmitting={isSubmitting}
        label={t("auth:merchant.signUp.submit")}
        submittingLabel={t("auth:merchant.signUp.submitting")}
        statusLabel={t("auth:merchant.signUp.status")}
      />
    </form>
  )
}
