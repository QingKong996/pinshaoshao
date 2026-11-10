import { AlertCircle, RotateCw } from "lucide-react"
import { useTranslation } from "react-i18next"
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { describeMerchantError } from "@/lib/merchant-errors"
import { cn } from "@/lib/utils"

type QueryErrorProps = {
  error: unknown
  onRetry?: () => void
  className?: string
}

export function QueryError({ error, onRetry, className }: QueryErrorProps) {
  const { t } = useTranslation("merchant")

  return (
    <Alert variant="destructive" className={cn(className)}>
      <AlertCircle aria-hidden />
      <AlertTitle>{t("common.loadError")}</AlertTitle>
      <AlertDescription>
        {describeMerchantError(error, t)}
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
