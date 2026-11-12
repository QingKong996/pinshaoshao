import { useTranslation } from "react-i18next"
import { QueryError as SharedQueryError } from "@/shared/ui/query-error"
import { describeDashboardError } from "../model/errors"

export function QueryError({
  error,
  onRetry,
}: {
  error: unknown
  onRetry?: () => void
}) {
  const { t } = useTranslation("console")
  return (
    <SharedQueryError
      title={t("common.loadError")}
      description={describeDashboardError(error, t)}
      retryLabel={t("common.retry")}
      onRetry={onRetry}
    />
  )
}
