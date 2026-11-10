import { useTranslation } from "react-i18next"
import { Badge } from "@/components/ui/badge"

type SpecBadgesProps = {
  specs: Record<string, string>
}

export function SpecBadges({ specs }: SpecBadgesProps) {
  const { t } = useTranslation("merchant")
  const entries = Object.entries(specs)

  if (entries.length === 0) {
    return <span className="text-muted-foreground">{t("common.none")}</span>
  }

  return (
    <span className="flex flex-wrap items-center gap-1">
      {entries.map(([key, value]) => (
        <Badge key={key} variant="outline">
          {`${key}: ${value}`}
        </Badge>
      ))}
    </span>
  )
}
