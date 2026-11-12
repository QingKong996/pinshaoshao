import { useTranslation } from "react-i18next"
import { SiteFooter } from "@/components/site-footer"
import { SiteHeader } from "@/components/site-header"
import { useDocumentTitle } from "@/hooks/use-document-title"

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
