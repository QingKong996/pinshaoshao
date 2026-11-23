import { useTranslation } from "react-i18next"
import { Package, ShoppingCart } from "lucide-react"
import { Link } from "react-router"
import { BrandMark } from "@/shared/ui/brand-mark"
import { LanguageSwitcher } from "@/shared/ui/language-switcher"
import { UserMenu, useAuth } from "@/features/auth"
import { buttonVariants } from "@/shared/ui/button"

export function SiteHeader() {
  const { session, signOut } = useAuth()
  const { t } = useTranslation()

  return (
    <header className="sticky top-0 z-50 border-b border-border/60 bg-background/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center gap-4 px-4 sm:px-6">
        <BrandMark />
        <div className="ml-auto flex items-center gap-2">
          {session ? (
            <>
              <Link
                to="/orders"
                aria-label={t("common:nav.orders")}
                title={t("common:nav.orders")}
                className={buttonVariants({
                  variant: "ghost",
                  size: "icon-sm",
                })}
              >
                <Package aria-hidden />
              </Link>
              <Link
                to="/cart"
                aria-label={t("common:nav.cart")}
                title={t("common:nav.cart")}
                className={buttonVariants({
                  variant: "ghost",
                  size: "icon-sm",
                })}
              >
                <ShoppingCart aria-hidden />
              </Link>
              <UserMenu
                session={session}
                signOut={signOut}
              />
            </>
          ) : (
            <Link
              to="/signin"
              className={buttonVariants({
                size: "sm",
              })}
            >
              {t("common:auth.signInOrSignUp")}
            </Link>
          )}
          <LanguageSwitcher />
        </div>
      </div>
    </header>
  )
}
