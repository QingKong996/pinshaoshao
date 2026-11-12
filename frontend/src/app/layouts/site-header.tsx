import { Search } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { BrandMark } from "@/shared/ui/brand-mark"
import { LanguageSwitcher } from "@/shared/ui/language-switcher"
import { UserMenu, useAuth } from "@/features/auth"
import { buttonVariants } from "@/shared/ui/button"
import { Input } from "@/shared/ui/input"

export function SiteHeader() {
  const { session } = useAuth()
  const { t } = useTranslation()

  return (
    <header className="sticky top-0 z-50 border-b border-border/60 bg-background/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between gap-4 px-4 sm:px-6 lg:grid lg:grid-cols-12">
        <BrandMark className="lg:col-span-3" />

        <div className="hidden min-w-0 flex-1 justify-center md:flex lg:col-span-6 lg:col-start-4">
          <div className="relative w-full max-w-xl">
            <Search
              aria-hidden
              className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              type="search"
              aria-label={t("common:nav.search")}
              placeholder={t("common:nav.search")}
              className="h-9 pl-8"
            />
          </div>
        </div>

        <div className="ml-auto flex items-center gap-2 lg:contents">
          {session ? (
            <UserMenu
              session={session}
              className="lg:col-span-1 lg:col-start-11 lg:p-0"
            />
          ) : (
            <Link
              to="/signin"
              className={buttonVariants({
                size: "sm",
                className: "lg:col-span-1 lg:col-start-11 lg:justify-self-end",
              })}
            >
              {t("common:auth.signInOrSignUp")}
            </Link>
          )}
          <LanguageSwitcher className="lg:col-span-1 lg:col-start-12 lg:justify-self-end" />
        </div>
      </div>
    </header>
  )
}
