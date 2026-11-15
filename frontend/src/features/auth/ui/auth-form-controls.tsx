import type { ReactNode } from "react"
import { LoaderCircle } from "lucide-react"
import type { UseFormRegisterReturn } from "react-hook-form"
import { Button } from "@/shared/ui/button"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import { PasswordInput } from "@/shared/ui/password-input"

type AuthTextFieldProps = {
  id: string
  label: string
  registerProps: UseFormRegisterReturn
  error?: string
  hint?: ReactNode
  type?: "text" | "email" | "password"
  autoComplete?: string
  autoFocus?: boolean
  autoCapitalize?: string
  spellCheck?: boolean
  required?: boolean
}

export function AuthTextField({
  id,
  label,
  registerProps,
  error,
  hint,
  type = "text",
  autoComplete,
  autoFocus,
  autoCapitalize,
  spellCheck,
  required = true,
}: AuthTextFieldProps) {
  const hintId = hint ? `${id}-hint` : undefined
  const errorId = error ? `${id}-error` : undefined
  const describedBy = [hintId, errorId].filter(Boolean).join(" ") || undefined
  const inputProps = {
    id,
    autoComplete,
    autoCapitalize,
    autoFocus,
    spellCheck,
    required,
    "aria-invalid": error ? true : undefined,
    "aria-describedby": describedBy,
    className: "h-10",
    ...registerProps,
  }

  return (
    <Field data-invalid={error ? true : undefined}>
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      {type === "password" ? (
        <PasswordInput {...inputProps} />
      ) : (
        <Input type={type} {...inputProps} />
      )}
      {hint ? <FieldDescription id={hintId}>{hint}</FieldDescription> : null}
      {error ? <FieldError id={errorId}>{error}</FieldError> : null}
    </Field>
  )
}

type AuthSubmitButtonProps = {
  isSubmitting: boolean
  label: string
  submittingLabel: string
  statusLabel: string
}

export function AuthSubmitButton({
  isSubmitting,
  label,
  submittingLabel,
  statusLabel,
}: AuthSubmitButtonProps) {
  return (
    <>
      <Button
        type="submit"
        size="lg"
        disabled={isSubmitting}
        className="h-10 w-full gap-2"
      >
        {isSubmitting ? (
          <>
            <LoaderCircle className="animate-spin" aria-hidden />
            {submittingLabel}
          </>
        ) : (
          label
        )}
      </Button>
      <span role="status" className="sr-only">
        {isSubmitting ? statusLabel : ""}
      </span>
    </>
  )
}
