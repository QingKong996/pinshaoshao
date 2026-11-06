import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { BrandMark } from "@/components/brand-mark"
import { buttonVariants } from "@/components/ui/button"
import { useDocumentTitle } from "@/lib/use-document-title"

export default function NotFoundPage() {
  const { t } = useTranslation()
  useDocumentTitle(t("common:notFound.documentTitle"))

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 bg-muted/30 px-4 text-center">
      <BrandMark />
      <div className="space-y-2">
        <p className="text-6xl font-semibold tracking-tight">404</p>
        <p className="text-sm text-muted-foreground">
          {t("common:notFound.description")}
        </p>
      </div>
      <Link to="/" className={buttonVariants()}>
        {t("common:notFound.backHome")}
      </Link>
    </div>
  )
}
