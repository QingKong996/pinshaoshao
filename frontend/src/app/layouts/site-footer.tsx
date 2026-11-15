import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { BrandMark } from "@/shared/ui/brand-mark"

const CURRENT_YEAR = new Date().getFullYear()

export function SiteFooter() {
  const { t } = useTranslation()

  return (
    <footer className="border-t border-border/60 bg-muted/30">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 px-4 py-10 sm:px-6 md:flex-row md:items-start md:justify-between">
        <div className="space-y-4">
          <BrandMark />
          <p className="max-w-sm text-sm leading-relaxed text-muted-foreground">
            {t("common:footer.tagline")}
          </p>
          <p className="max-w-sm text-xs leading-relaxed text-muted-foreground/70">
            {t("common:footer.disclaimer")}
          </p>
        </div>
        <Link
          to="/merchant"
          className="w-fit rounded-sm text-sm text-muted-foreground underline-offset-4 transition-colors hover:text-foreground hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
        >
          {t("common:footer.join.merchants")}
        </Link>
      </div>

      <div className="border-t border-border/60">
        <div className="mx-auto w-full max-w-6xl px-4 py-5 text-xs text-muted-foreground sm:px-6">
          {t("common:footer.copyright", { year: CURRENT_YEAR })}
        </div>
      </div>
    </footer>
  )
}
