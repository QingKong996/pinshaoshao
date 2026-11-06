import { useTranslation } from "react-i18next"
import { BrandMark } from "@/components/brand-mark"

const FOOTER_SECTIONS = [
  {
    titleKey: "common:footer.exploreTitle",
    className: "md:col-span-1 md:col-start-8",
    linkKeys: ["common:nav.features", "common:nav.howItWorks"],
  },
  {
    titleKey: "common:footer.joinTitle",
    className: "md:col-span-1 md:col-start-10",
    linkKeys: [
      "common:footer.join.merchants",
      "common:footer.join.partnership",
      "common:footer.join.contact",
    ],
  },
  {
    titleKey: "common:footer.helpTitle",
    className: "md:col-span-1 md:col-start-12",
    linkKeys: [
      "common:footer.help.faq",
      "common:footer.help.shipping",
      "common:footer.help.afterSales",
    ],
  },
] as const

const CURRENT_YEAR = new Date().getFullYear()

export function SiteFooter() {
  const { t } = useTranslation()

  return (
    <footer className="border-t border-border/60 bg-muted/30">
      <div className="mx-auto grid w-full max-w-6xl gap-10 px-4 py-12 sm:px-6 md:grid-cols-12 md:gap-4">
        <div className="space-y-4 md:col-span-7">
          <BrandMark />
          <p className="max-w-sm text-sm leading-relaxed text-muted-foreground">
            {t("common:footer.tagline")}
          </p>
          <p className="max-w-sm text-xs leading-relaxed text-muted-foreground/70">
            {t("common:footer.disclaimer")}
          </p>
        </div>

        {FOOTER_SECTIONS.map((section) => (
          <div key={section.titleKey} className={section.className}>
            <h3 className="mb-3 text-sm font-medium">{t(section.titleKey)}</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              {section.linkKeys.map((labelKey) => (
                <li key={labelKey}>
                  <button
                    type="button"
                    className="cursor-pointer transition-colors hover:text-foreground"
                  >
                    {t(labelKey)}
                  </button>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>

      <div className="border-t border-border/60">
        <div className="mx-auto w-full max-w-6xl px-4 py-5 text-xs text-muted-foreground sm:px-6">
          {t("common:footer.copyright", { year: CURRENT_YEAR })}
        </div>
      </div>
    </footer>
  )
}
