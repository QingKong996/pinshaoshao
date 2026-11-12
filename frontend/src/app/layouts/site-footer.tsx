import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { BrandMark } from "@/shared/ui/brand-mark"

const FOOTER_SECTIONS = [
  {
    titleKey: "common:footer.exploreTitle",
    className: "md:col-span-1 md:col-start-8",
    links: [
      { labelKey: "common:nav.features" },
      { labelKey: "common:nav.howItWorks" },
    ],
  },
  {
    titleKey: "common:footer.joinTitle",
    className: "md:col-span-1 md:col-start-10",
    links: [
      { labelKey: "common:footer.join.merchants", to: "/merchant/signup" },
      { labelKey: "common:footer.join.partnership" },
      { labelKey: "common:footer.join.contact" },
    ],
  },
  {
    titleKey: "common:footer.helpTitle",
    className: "md:col-span-1 md:col-start-12",
    links: [
      { labelKey: "common:footer.help.faq" },
      { labelKey: "common:footer.help.shipping" },
      { labelKey: "common:footer.help.afterSales" },
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
              {section.links.map((link) => (
                <li key={link.labelKey}>
                  {"to" in link ? (
                    <Link
                      to={link.to}
                      className="transition-colors hover:text-foreground"
                    >
                      {t(link.labelKey)}
                    </Link>
                  ) : (
                    <button
                      type="button"
                      className="cursor-pointer transition-colors hover:text-foreground"
                    >
                      {t(link.labelKey)}
                    </button>
                  )}
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
