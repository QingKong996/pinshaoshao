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
import { describeUserAuthError } from "../model/user-errors"
import { useAuth } from "../model/user-context"

const INITIAL_VALUES: SignInValues = {
  username: "",
  password: "",
}

export function SignInForm() {
  const { t } = useTranslation()
  const { session, signIn } = useAuth()
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
    return <Navigate to="/" replace />
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
          message={describeUserAuthError(serverError, "login", t)}
        />
      ) : null}

      <FieldGroup>
        <AuthTextField
          id="signin-username"
          label={t("auth:fields.username")}
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
          id="signin-password"
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
        label={t("auth:signIn.submit")}
        submittingLabel={t("auth:signIn.submitting")}
        statusLabel={t("auth:signIn.status")}
      />
    </form>
  )
}
