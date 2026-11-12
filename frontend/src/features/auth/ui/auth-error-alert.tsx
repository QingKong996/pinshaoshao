import { CircleAlert } from "lucide-react"
import { Alert, AlertDescription } from "@/shared/ui/alert"

type AuthErrorAlertProps = {
  message: string
}

export function AuthErrorAlert({ message }: AuthErrorAlertProps) {
  return (
    <Alert
      variant="destructive"
      className="items-start has-[>svg]:gap-x-2.5 rounded-lg border-destructive/25 bg-destructive/5 px-3 py-2.5"
    >
      <CircleAlert className="size-4 shrink-0" aria-hidden />
      <AlertDescription className="leading-relaxed">
        {message}
      </AlertDescription>
    </Alert>
  )
}
