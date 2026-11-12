import { useTranslation } from "react-i18next"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

export default function HomePage() {
  const { t } = useTranslation()
  useDocumentTitle(t("home:documentTitle"))

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="flex-1" />
      <SiteFooter />
    </div>
  )
}
