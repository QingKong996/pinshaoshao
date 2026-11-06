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
import { useAuth } from "@/lib/auth-context"
import { describeAuthError } from "@/lib/auth-errors"
import {
  signUpSchema,
  translateAuthValidationMessage,
  type SignUpValues,
} from "@/lib/auth-schemas"

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
          message={describeAuthError(serverError, "register", t)}
        />
      ) : null}

      <FieldGroup>
        <Field data-invalid={errors.username ? true : undefined}>
          <FieldLabel htmlFor="signup-username">
            {t("auth:fields.username")}
          </FieldLabel>
          <Input
            id="signup-username"
            autoComplete="username"
            autoCapitalize="none"
            autoFocus
            spellCheck={false}
            required
            aria-invalid={errors.username ? true : undefined}
            aria-describedby={
              errors.username
                ? "signup-username-hint signup-username-error"
                : "signup-username-hint"
            }
            className="h-10"
            {...register("username", { onChange: clearServerError })}
          />
          <FieldDescription id="signup-username-hint">
            {t("auth:hints.username")}
          </FieldDescription>
          {errors.username ? (
            <FieldError id="signup-username-error">
              {translateAuthValidationMessage(errors.username.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.email ? true : undefined}>
          <FieldLabel htmlFor="signup-email">{t("auth:fields.email")}</FieldLabel>
          <Input
            id="signup-email"
            type="email"
            autoComplete="email"
            autoCapitalize="none"
            spellCheck={false}
            aria-invalid={errors.email ? true : undefined}
            aria-describedby={errors.email ? "signup-email-error" : undefined}
            className="h-10"
            {...register("email", { onChange: clearServerError })}
          />
          <FieldDescription>{t("auth:hints.optional")}</FieldDescription>
          {errors.email ? (
            <FieldError id="signup-email-error">
              {translateAuthValidationMessage(errors.email.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.password ? true : undefined}>
          <FieldLabel htmlFor="signup-password">
            {t("auth:fields.password")}
          </FieldLabel>
          <PasswordInput
            id="signup-password"
            autoComplete="new-password"
            required
            aria-invalid={errors.password ? true : undefined}
            aria-describedby={
              errors.password
                ? "signup-password-hint signup-password-error"
                : "signup-password-hint"
            }
            className="h-10"
            {...register("password", {
              deps: ["confirmPassword"],
              onChange: clearServerError,
            })}
          />
          <FieldDescription id="signup-password-hint">
            {t("auth:hints.password")}
          </FieldDescription>
          {errors.password ? (
            <FieldError id="signup-password-error">
              {translateAuthValidationMessage(errors.password.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <Field data-invalid={errors.confirmPassword ? true : undefined}>
          <FieldLabel htmlFor="signup-confirm-password">
            {t("auth:fields.confirmPassword")}
          </FieldLabel>
          <PasswordInput
            id="signup-confirm-password"
            autoComplete="new-password"
            required
            aria-invalid={errors.confirmPassword ? true : undefined}
            aria-describedby={
              errors.confirmPassword
                ? "signup-confirm-password-error"
                : undefined
            }
            className="h-10"
            {...register("confirmPassword", { onChange: clearServerError })}
          />
          {errors.confirmPassword ? (
            <FieldError id="signup-confirm-password-error">
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
            {t("auth:signUp.submitting")}
          </>
        ) : (
          t("auth:signUp.submit")
        )}
      </Button>

      <span role="status" className="sr-only">
        {isSubmitting ? t("auth:signUp.status") : ""}
      </span>
    </form>
  )
}
