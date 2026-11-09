import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle } from "lucide-react"
import { useForm, type Resolver } from "react-hook-form"
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
  useAuth,
  useMerchantAuth,
  type AuthKind,
  type AuthSession,
} from "@/lib/auth-context"
import { describeAuthError } from "@/lib/auth-errors"
import {
  merchantSignUpSchema,
  signUpSchema,
  translateAuthValidationMessage,
  type SignUpValues,
} from "@/lib/auth-schemas"

const INITIAL_VALUES: SignUpFormValues = {
  name: "",
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
}

const COPY = {
  user: {
    username: "auth:fields.username",
    emailHint: "auth:hints.optional",
    submit: "auth:signUp.submit",
    submitting: "auth:signUp.submitting",
    status: "auth:signUp.status",
  },
  merchant: {
    username: "auth:merchant.fields.username",
    emailHint: "auth:merchant.hints.email",
    submit: "auth:merchant.signUp.submit",
    submitting: "auth:merchant.signUp.submitting",
    status: "auth:merchant.signUp.status",
  },
} as const

type SignUpFormValues = SignUpValues & { name?: string }

const RESOLVERS = {
  user: zodResolver(signUpSchema),
  merchant: zodResolver(merchantSignUpSchema),
} as const

type SignUpFormViewProps = {
  kind: AuthKind
  session: AuthSession | null
  onSignUp: (values: SignUpFormValues) => Promise<void>
}

function SignUpFormView({ kind, session, onSignUp }: SignUpFormViewProps) {
  const { t } = useTranslation()
  const copy = COPY[kind]
  const [serverError, setServerError] = useState<unknown>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignUpFormValues>({
    resolver: RESOLVERS[kind] as Resolver<SignUpFormValues>,
    defaultValues: INITIAL_VALUES,
  })

  async function onSubmit(values: SignUpFormValues) {
    setServerError(null)
    try {
      await onSignUp(values)
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
        {kind === "merchant" ? (
          <Field data-invalid={errors.name ? true : undefined}>
            <FieldLabel htmlFor="signup-shop-name">
              {t("auth:merchant.fields.name")}
            </FieldLabel>
            <Input
              id="signup-shop-name"
              autoComplete="organization"
              autoFocus
              required
              aria-invalid={errors.name ? true : undefined}
              aria-describedby={
                errors.name
                  ? "signup-shop-name-hint signup-shop-name-error"
                  : "signup-shop-name-hint"
              }
              className="h-10"
              {...register("name", { onChange: clearServerError })}
            />
            <FieldDescription id="signup-shop-name-hint">
              {t("auth:merchant.hints.name")}
            </FieldDescription>
            {errors.name ? (
              <FieldError id="signup-shop-name-error">
                {translateAuthValidationMessage(errors.name.message, t)}
              </FieldError>
            ) : null}
          </Field>
        ) : null}

        <Field data-invalid={errors.username ? true : undefined}>
          <FieldLabel htmlFor="signup-username">{t(copy.username)}</FieldLabel>
          <Input
            id="signup-username"
            autoComplete="username"
            autoCapitalize="none"
            autoFocus={kind === "user"}
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
            required={kind === "merchant"}
            aria-invalid={errors.email ? true : undefined}
            aria-describedby={errors.email ? "signup-email-error" : undefined}
            className="h-10"
            {...register("email", { onChange: clearServerError })}
          />
          <FieldDescription>{t(copy.emailHint)}</FieldDescription>
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

export function SignUpForm() {
  const { session, signUp } = useAuth()

  return (
    <SignUpFormView
      kind="user"
      session={session}
      onSignUp={(values) =>
        signUp({
          username: values.username,
          email: values.email || undefined,
          password: values.password,
        })
      }
    />
  )
}

export function MerchantSignUpForm() {
  const { session, signUp } = useMerchantAuth()

  return (
    <SignUpFormView
      kind="merchant"
      session={session}
      onSignUp={(values) =>
        signUp({
          username: values.username,
          name: values.name ?? "",
          email: values.email,
          password: values.password,
        })
      }
    />
  )
}
