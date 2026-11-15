import { ArrowRight } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { useMerchantAuth } from "@/features/auth"
import { BrandMark } from "@/shared/ui/brand-mark"
import { buttonVariants } from "@/shared/ui/button"
import { LanguageSwitcher } from "@/shared/ui/language-switcher"

const NAV_ITEMS = [
  {
    href: "#merchant-features",
    labelKey: "home:merchant.header.features",
  },
  {
    href: "#merchant-how-it-works",
    labelKey: "home:merchant.header.howItWorks",
  },
] as const

const NAV_LINK_CLASS =
  "rounded-sm transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"

export function MerchantSiteHeader() {
  const { session } = useMerchantAuth()
  const { t } = useTranslation()

  return (
    <header className="sticky top-0 z-50 border-b border-border/60 bg-background/85 backdrop-blur-xl">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center gap-3 px-4 sm:px-6">
        <BrandMark to="/merchant" />
        <span className="hidden rounded-full border border-border/70 bg-muted/50 px-2.5 py-1 text-[11px] font-medium text-muted-foreground sm:inline-flex">
          {t("home:merchant.header.platform")}
        </span>

        <nav className="ml-4 hidden items-center gap-5 text-sm text-muted-foreground lg:flex">
          {NAV_ITEMS.map((item) => (
            <a
              key={item.href}
              href={item.href}
              className={NAV_LINK_CLASS}
            >
              {t(item.labelKey)}
            </a>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <LanguageSwitcher className="hidden sm:flex" />
          {session ? (
            <Link
              to="/merchant/dashboard"
              className={buttonVariants({ size: "sm" })}
            >
              {t("home:merchant.header.console")}
              <ArrowRight aria-hidden />
            </Link>
          ) : (
            <>
              <Link
                to="/merchant/signin"
                className={buttonVariants({
                  variant: "ghost",
                  size: "sm",
                  className: "hidden sm:inline-flex",
                })}
              >
                {t("home:merchant.header.signIn")}
              </Link>
              <Link
                to="/merchant/signup"
                className={buttonVariants({ size: "sm" })}
              >
                {t("home:merchant.header.signUp")}
                <ArrowRight aria-hidden />
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
