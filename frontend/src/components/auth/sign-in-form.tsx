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
import type { LoginRequest } from "@/lib/api"
import {
  useAuth,
  useMerchantAuth,
  type AuthKind,
  type AuthSession,
} from "@/lib/auth-context"
import { describeAuthError } from "@/lib/auth-errors"
import {
  signInSchema,
  translateAuthValidationMessage,
  type SignInValues,
} from "@/lib/auth-schemas"

const INITIAL_VALUES: SignInValues = {
  username: "",
  password: "",
}

const COPY = {
  user: {
    username: "auth:fields.username",
    submit: "auth:signIn.submit",
    submitting: "auth:signIn.submitting",
    status: "auth:signIn.status",
  },
  merchant: {
    username: "auth:merchant.fields.username",
    submit: "auth:merchant.signIn.submit",
    submitting: "auth:merchant.signIn.submitting",
    status: "auth:merchant.signIn.status",
  },
} as const

type SignInFormViewProps = {
  kind: AuthKind
  session: AuthSession | null
  onSignIn: (credentials: LoginRequest) => Promise<void>
}

function SignInFormView({ kind, session, onSignIn }: SignInFormViewProps) {
  const { t } = useTranslation()
  const copy = COPY[kind]
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
      await onSignIn(values)
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
        <AuthErrorAlert message={describeAuthError(serverError, "login", t)} />
      ) : null}

      <FieldGroup>
        <Field data-invalid={errors.username ? true : undefined}>
          <FieldLabel htmlFor="signin-username">{t(copy.username)}</FieldLabel>
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
            {t(copy.submitting)}
          </>
        ) : (
          t(copy.submit)
        )}
      </Button>

      <span role="status" className="sr-only">
        {isSubmitting ? t(copy.status) : ""}
      </span>
    </form>
  )
}

export function SignInForm() {
  const { session, signIn } = useAuth()

  return <SignInFormView kind="user" session={session} onSignIn={signIn} />
}

export function MerchantSignInForm() {
  const { session, signIn } = useMerchantAuth()

  return (
    <SignInFormView kind="merchant" session={session} onSignIn={signIn} />
  )
}
