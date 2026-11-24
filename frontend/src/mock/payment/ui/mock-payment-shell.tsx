import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Badge } from "@/shared/ui/badge"
import { BrandMark } from "@/shared/ui/brand-mark"
import { LanguageSwitcher } from "@/shared/ui/language-switcher"
import { MOCK_PAYMENT_ROUTE_PATHS } from "../model/paths"

export function MockPaymentShell({ children }: { children: ReactNode }) {
  const { t } = useTranslation("mockpayment")

  return (
    <div className="flex min-h-svh flex-col bg-muted/20">
      <header className="border-b border-border/70 bg-background/90 backdrop-blur-xl">
        <div className="mx-auto flex h-16 w-full max-w-6xl items-center gap-3 px-4 sm:px-6">
          <BrandMark />
          <span
            className="hidden h-6 w-px bg-border sm:block"
            aria-hidden
          />
          <Link
            to={MOCK_PAYMENT_ROUTE_PATHS.list}
            className="flex min-w-0 flex-col rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
          >
            <span className="truncate text-sm font-semibold tracking-tight">
              {t("shell.title")}
            </span>
            <span className="hidden truncate text-[10px] font-medium tracking-[0.16em] text-muted-foreground uppercase sm:block">
              {t("shell.subtitle")}
            </span>
          </Link>
          <div className="ml-auto flex items-center gap-2">
            <Badge variant="outline" className="hidden sm:inline-flex">
              {t("shell.sandbox")}
            </Badge>
            <LanguageSwitcher />
          </div>
        </div>
      </header>

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6 sm:py-12">
        {children}
      </main>

      <footer className="border-t border-border/70 bg-background/60">
        <div className="mx-auto w-full max-w-6xl px-4 py-5 text-xs text-muted-foreground sm:px-6">
          {t("shell.footer")}
        </div>
      </footer>
    </div>
  )
}
