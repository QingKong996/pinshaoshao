import { useTranslation } from "react-i18next"
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { FEATURES } from "./merchant-home-content"
import { MerchantSectionIntro } from "./merchant-section-intro"

export function MerchantFeatures() {
  const { t } = useTranslation()

  return (
    <section
      id="merchant-features"
      className="border-y border-border/60 bg-muted/20 py-20 lg:py-24"
    >
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <MerchantSectionIntro
          eyebrow={t("home:merchant.features.eyebrow")}
          title={t("home:merchant.features.title")}
          description={t("home:merchant.features.description")}
        />

        <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {FEATURES.map(({ key, icon: Icon, titleKey, descriptionKey }) => (
            <Card
              key={key}
              className="rounded-2xl border border-border/70 bg-background/80 shadow-none ring-0 transition-colors hover:border-orange-500/25"
            >
              <CardHeader>
                <span className="mb-2 grid size-10 place-items-center rounded-xl bg-orange-500/10 text-orange-600 ring-1 ring-orange-500/10 dark:text-orange-400">
                  <Icon className="size-4" aria-hidden />
                </span>
                <CardTitle>{t(titleKey)}</CardTitle>
                <CardDescription className="leading-relaxed">
                  {t(descriptionKey)}
                </CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>
      </div>
    </section>
  )
}
