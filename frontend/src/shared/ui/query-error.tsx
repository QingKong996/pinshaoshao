import { AlertCircle, RotateCw } from "lucide-react"
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"

type QueryErrorProps = {
  title: string
  description: string
  retryLabel: string
  onRetry?: () => void
}

export function QueryError({
  title,
  description,
  retryLabel,
  onRetry,
}: QueryErrorProps) {
  return (
    <Alert variant="destructive">
      <AlertCircle aria-hidden />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>
        {description}
      </AlertDescription>
      {onRetry ? (
        <AlertAction>
          <Button variant="outline" size="sm" onClick={onRetry}>
            <RotateCw aria-hidden />
            {retryLabel}
          </Button>
        </AlertAction>
      ) : null}
    </Alert>
  )
}
