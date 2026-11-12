import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router"
import { AuthErrorAlert } from "./auth-error-alert"
import { PasswordInput } from "@/shared/ui/password-input"
import { Button } from "@/shared/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
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
        <Field data-invalid={errors.username ? true : undefined}>
          <FieldLabel htmlFor="merchant-signin-username">
            {t("auth:merchant.fields.username")}
          </FieldLabel>
          <Input
            id="merchant-signin-username"
            autoComplete="username"
            autoCapitalize="none"
            autoFocus
            spellCheck={false}
            required
            aria-invalid={errors.username ? true : undefined}
            aria-describedby={
              errors.username ? "merchant-signin-username-error" : undefined
            }
            className="h-10"
            {...register("username", { onChange: clearServerError })}
          />
          {errors.username ? (
            <FieldError id="merchant-signin-username-error">
              {translateAuthValidationMessage(errors.username.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.password ? true : undefined}>
          <FieldLabel htmlFor="merchant-signin-password">
            {t("auth:fields.password")}
          </FieldLabel>
          <PasswordInput
            id="merchant-signin-password"
            autoComplete="current-password"
            required
            aria-invalid={errors.password ? true : undefined}
            aria-describedby={
              errors.password ? "merchant-signin-password-error" : undefined
            }
            className="h-10"
            {...register("password", { onChange: clearServerError })}
          />
          {errors.password ? (
            <FieldError id="merchant-signin-password-error">
              {translateAuthValidationMessage(errors.password.message, t)}
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
            {t("auth:merchant.signIn.submitting")}
          </>
        ) : (
          t("auth:merchant.signIn.submit")
        )}
      </Button>

      <span role="status" className="sr-only">
        {isSubmitting ? t("auth:merchant.signIn.status") : ""}
      </span>
    </form>
  )
}
