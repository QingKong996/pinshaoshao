import { useTranslation } from "react-i18next"
import { MerchantSiteFooter } from "../../layouts/merchant-site-footer"
import { MerchantSiteHeader } from "../../layouts/merchant-site-header"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { MerchantClosingCta } from "./merchant-closing-cta"
import { MerchantFeatures } from "./merchant-features"
import { MerchantHero } from "./merchant-hero"
import { MerchantHowItWorks } from "./merchant-how-it-works"

export default function MerchantHomePage() {
  const { t } = useTranslation()
  useDocumentTitle(t("home:merchant.documentTitle"))

  return (
    <div className="flex min-h-svh flex-col">
      <MerchantSiteHeader />
      <main className="flex-1">
        <MerchantHero />
        <MerchantFeatures />
        <MerchantHowItWorks />
        <MerchantClosingCta />
      </main>
      <MerchantSiteFooter />
    </div>
  )
}
