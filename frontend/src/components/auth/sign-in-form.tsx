import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router"
import { AuthErrorAlert } from "@/components/auth/auth-error-alert"
import { PasswordInput } from "@/components/auth/password-input"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  signInSchema,
  translateAuthValidationMessage,
  type SignInValues,
} from "@/lib/auth-schemas"
import { describeUserAuthError } from "@/lib/user-auth-errors"
import { useAuth } from "@/lib/user-auth-context"

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
        <Field data-invalid={errors.username ? true : undefined}>
          <FieldLabel htmlFor="signin-username">
            {t("auth:fields.username")}
          </FieldLabel>
          <Input
            id="signin-username"
            autoComplete="username"
            autoCapitalize="none"
            autoFocus
            spellCheck={false}
            required
            aria-invalid={errors.username ? true : undefined}
            aria-describedby={
              errors.username ? "signin-username-error" : undefined
            }
            className="h-10"
            {...register("username", { onChange: clearServerError })}
          />
          {errors.username ? (
            <FieldError id="signin-username-error">
              {translateAuthValidationMessage(errors.username.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.password ? true : undefined}>
          <FieldLabel htmlFor="signin-password">
            {t("auth:fields.password")}
          </FieldLabel>
          <PasswordInput
            id="signin-password"
            autoComplete="current-password"
            required
            aria-invalid={errors.password ? true : undefined}
            aria-describedby={
              errors.password ? "signin-password-error" : undefined
            }
            className="h-10"
            {...register("password", { onChange: clearServerError })}
          />
          {errors.password ? (
            <FieldError id="signin-password-error">
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
            {t("auth:signIn.submitting")}
          </>
        ) : (
          t("auth:signIn.submit")
        )}
      </Button>

      <span role="status" className="sr-only">
        {isSubmitting ? t("auth:signIn.status") : ""}
      </span>
    </form>
  )
}
