import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { BrandMark } from "@/shared/ui/brand-mark"

const CURRENT_YEAR = new Date().getFullYear()
const LINK_CLASS =
  "rounded-sm transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"

const FOOTER_LINKS = [
  {
    labelKey: "home:merchant.footer.links.features",
    href: "#merchant-features",
  },
  {
    labelKey: "home:merchant.footer.links.howItWorks",
    href: "#merchant-how-it-works",
  },
] as const

export function MerchantSiteFooter() {
  const { t } = useTranslation()

  return (
    <footer className="border-t border-border/60 bg-muted/25">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 px-4 py-10 sm:px-6 md:flex-row md:items-start md:justify-between">
        <div className="max-w-sm space-y-4">
          <BrandMark to="/merchant" />
          <p className="text-sm leading-relaxed text-muted-foreground">
            {t("home:merchant.footer.tagline")}
          </p>
        </div>

        <div className="flex flex-wrap gap-x-6 gap-y-3 text-sm text-muted-foreground">
          {FOOTER_LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className={LINK_CLASS}
            >
              {t(link.labelKey)}
            </a>
          ))}
          <Link
            to="/merchant/signin"
            className={LINK_CLASS}
          >
            {t("home:merchant.footer.links.signIn")}
          </Link>
          <Link
            to="/merchant/signup"
            className={LINK_CLASS}
          >
            {t("home:merchant.footer.links.signUp")}
          </Link>
          <Link to="/" className={LINK_CLASS}>
            {t("home:merchant.footer.links.backToShop")}
          </Link>
        </div>
      </div>

      <div className="border-t border-border/60">
        <div className="mx-auto flex w-full max-w-6xl flex-col gap-2 px-4 py-5 text-xs text-muted-foreground sm:px-6 md:flex-row md:items-center md:justify-between">
          <span>
            {t("home:merchant.footer.copyright", { year: CURRENT_YEAR })}
          </span>
          <span>{t("home:merchant.footer.disclaimer")}</span>
        </div>
      </div>
    </footer>
  )
}
