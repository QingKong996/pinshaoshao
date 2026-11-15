import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router"
import { AuthErrorAlert } from "./auth-error-alert"
import { AuthSubmitButton, AuthTextField } from "./auth-form-controls"
import { FieldGroup } from "@/shared/ui/field"
import {
  signUpSchema,
  translateAuthValidationMessage,
  type SignUpValues,
} from "../model/schemas"
import { describeUserAuthError } from "../model/user-errors"
import { useAuth } from "../model/user-context"

const INITIAL_VALUES: SignUpValues = {
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
}

export function SignUpForm() {
  const { t } = useTranslation()
  const { session, signUp } = useAuth()
  const [serverError, setServerError] = useState<unknown>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignUpValues>({
    resolver: zodResolver(signUpSchema),
    defaultValues: INITIAL_VALUES,
  })

  async function onSubmit(values: SignUpValues) {
    setServerError(null)
    try {
      await signUp({
        username: values.username,
        email: values.email || undefined,
        password: values.password,
      })
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
          message={describeUserAuthError(serverError, "register", t)}
        />
      ) : null}

      <FieldGroup>
        <AuthTextField
          id="signup-username"
          label={t("auth:fields.username")}
          autoComplete="username"
          autoCapitalize="none"
          autoFocus
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
          id="signup-email"
          label={t("auth:fields.email")}
          type="email"
          autoComplete="email"
          autoCapitalize="none"
          spellCheck={false}
          required={false}
          hint={t("auth:hints.optional")}
          error={translateAuthValidationMessage(
            errors.email?.message,
            t,
          )}
          registerProps={register("email", {
            onChange: clearServerError,
          })}
        />
        <AuthTextField
          id="signup-password"
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
          id="signup-confirm-password"
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
        label={t("auth:signUp.submit")}
        submittingLabel={t("auth:signUp.submitting")}
        statusLabel={t("auth:signUp.status")}
      />
    </form>
  )
}
