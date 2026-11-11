import { AlertCircle, RotateCw } from "lucide-react"
import { useTranslation } from "react-i18next"
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { describeConsoleError } from "@/lib/merchant-console-errors"

type QueryErrorProps = {
  error: unknown
  onRetry?: () => void
}

export function QueryError({ error, onRetry }: QueryErrorProps) {
  const { t } = useTranslation("console")

  return (
    <Alert variant="destructive">
      <AlertCircle aria-hidden />
      <AlertTitle>{t("common.loadError")}</AlertTitle>
      <AlertDescription>
        {describeConsoleError(error, t)}
      </AlertDescription>
      {onRetry ? (
        <AlertAction>
          <Button variant="outline" size="sm" onClick={onRetry}>
            <RotateCw aria-hidden />
            {t("common.retry")}
          </Button>
        </AlertAction>
      ) : null}
    </Alert>
  )
}
