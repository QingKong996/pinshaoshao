import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router"
import { AuthErrorAlert } from "./auth-error-alert"
import { AuthSubmitButton, AuthTextField } from "./auth-form-controls"
import { FieldGroup } from "@/shared/ui/field"
import {
  signInSchema,
  translateAuthValidationMessage,
  type SignInValues,
} from "../model/schemas"
import { describeMerchantAuthError } from "../model/merchant-errors"
import { useMerchantAuth } from "../model/merchant-context"

const INITIAL_VALUES: SignInValues = {
  username: "",
  password: "",
}

export function MerchantSignInForm({ redirectTo }: { redirectTo: string }) {
  const { t } = useTranslation()
  const { session, signIn } = useMerchantAuth()
  const [serverError, setServerError] = useState<unknown>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignInValues>({
    resolver: zodResolver(signInSchema),
    defaultValues: INITIAL_VALUES,
  })

  async function onSubmit(values: SignInValues) {
    setServerError(null)
    try {
      await signIn(values)
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
          message={describeMerchantAuthError(serverError, "login", t)}
        />
      ) : null}

      <FieldGroup>
        <AuthTextField
          id="merchant-signin-username"
          label={t("auth:merchant.fields.username")}
          autoComplete="username"
          autoCapitalize="none"
          autoFocus
          spellCheck={false}
          error={translateAuthValidationMessage(
            errors.username?.message,
            t,
          )}
          registerProps={register("username", {
            onChange: clearServerError,
          })}
        />
        <AuthTextField
          id="merchant-signin-password"
          label={t("auth:fields.password")}
          type="password"
          autoComplete="current-password"
          error={translateAuthValidationMessage(
            errors.password?.message,
            t,
          )}
          registerProps={register("password", {
            onChange: clearServerError,
          })}
        />
      </FieldGroup>

      <AuthSubmitButton
        isSubmitting={isSubmitting}
        label={t("auth:merchant.signIn.submit")}
        submittingLabel={t("auth:merchant.signIn.submitting")}
        statusLabel={t("auth:merchant.signIn.status")}
      />
    </form>
  )
}
