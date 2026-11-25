import { useState, type FormEvent } from "react"
import { CircleAlert, LoaderCircle } from "lucide-react"
import { useLocation, useNavigate } from "react-router"
import { ApiError } from "@/shared/api/client"
import { Alert, AlertDescription } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import { Field, FieldGroup, FieldLabel } from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"

type MockSignInFormProps = {
  fallbackPath: string
  labels: {
    username: string
    password: string
    submit: string
    submitting: string
    invalidCredentials: string
    unavailable: string
  }
  signIn: (username: string, password: string) => Promise<void>
}

export function MockSignInForm({
  fallbackPath,
  labels,
  signIn,
}: MockSignInFormProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState("")

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError("")
    try {
      await signIn(username, password)
      const from = (location.state as { from?: unknown } | null)?.from
      navigate(typeof from === "string" ? from : fallbackPath, { replace: true })
    } catch (caught) {
      setError(
        caught instanceof ApiError && caught.status === 401
          ? labels.invalidCredentials
          : labels.unavailable,
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5">
      {error ? (
        <Alert
          variant="destructive"
          className="items-start has-[>svg]:gap-x-2.5 rounded-lg border-destructive/25 bg-destructive/5 px-3 py-2.5"
        >
          <CircleAlert className="size-4 shrink-0" aria-hidden />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="mock-auth-username">{labels.username}</FieldLabel>
          <Input
            id="mock-auth-username"
            name="username"
            autoComplete="username"
            autoCapitalize="none"
            spellCheck={false}
            required
            value={username}
            onChange={(event) => setUsername(event.target.value)}
          />
        </Field>
        <Field>
          <FieldLabel htmlFor="mock-auth-password">{labels.password}</FieldLabel>
          <Input
            id="mock-auth-password"
            name="password"
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </Field>
      </FieldGroup>

      <Button type="submit" size="lg" disabled={submitting} className="h-10">
        {submitting ? (
          <>
            <LoaderCircle className="animate-spin" aria-hidden />
            {labels.submitting}
          </>
        ) : (
          labels.submit
        )}
      </Button>
    </form>
  )
}
