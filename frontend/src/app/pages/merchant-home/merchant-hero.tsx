import { ArrowRight, Check, Store } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Badge } from "@/shared/ui/badge"
import { buttonVariants } from "@/shared/ui/button"
import { MerchantConsolePreview } from "./merchant-console-preview"
import { TRUST_POINTS } from "./merchant-home-content"

export function MerchantHero() {
  const { t } = useTranslation()

  return (
    <section className="relative isolate overflow-hidden">
      <div aria-hidden className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute -top-48 left-1/2 h-[34rem] w-[62rem] -translate-x-1/2 rounded-full bg-[radial-gradient(ellipse_at_center,rgba(249,115,22,0.16),rgba(244,63,94,0.08)_42%,transparent_70%)]" />
        <div className="absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:72px_72px] opacity-45 [mask-image:linear-gradient(to_bottom,black,transparent_82%)]" />
      </div>

      <div className="mx-auto grid w-full max-w-6xl gap-14 px-4 py-16 sm:px-6 sm:py-20 lg:grid-cols-[1.02fr_0.98fr] lg:items-center lg:gap-16 lg:py-28">
        <div className="max-w-2xl">
          <Badge
            variant="outline"
            className="h-7 gap-1.5 border-orange-500/20 bg-orange-500/10 px-3 text-orange-700 dark:text-orange-300"
          >
            <Store aria-hidden />
            {t("home:merchant.hero.badge")}
          </Badge>

          <h1 className="mt-6 font-heading text-4xl font-semibold tracking-[-0.045em] text-balance sm:text-5xl lg:text-[3.6rem] lg:leading-[1.04]">
            {t("home:merchant.hero.title.lead")}
            <span className="mt-1 block bg-gradient-to-r from-orange-600 via-rose-500 to-rose-600 bg-clip-text text-transparent">
              {t("home:merchant.hero.title.highlight")}
            </span>
          </h1>

          <p className="mt-6 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg sm:leading-8">
            {t("home:merchant.hero.description")}
          </p>

          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Link
              to="/merchant/signup"
              className={buttonVariants({
                size: "lg",
                className: "h-11 rounded-xl px-5",
              })}
            >
              {t("home:merchant.hero.actions.signUp")}
              <ArrowRight aria-hidden />
            </Link>
            <Link
              to="/merchant/signin"
              className={buttonVariants({
                variant: "outline",
                size: "lg",
                className: "h-11 rounded-xl px-5",
              })}
            >
              {t("home:merchant.hero.actions.signIn")}
            </Link>
          </div>

          <ul className="mt-8 flex flex-wrap gap-x-5 gap-y-3">
            {TRUST_POINTS.map((key) => (
              <li
                key={key}
                className="flex items-center gap-2 text-sm text-muted-foreground"
              >
                <span className="grid size-5 place-items-center rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                  <Check className="size-3" aria-hidden />
                </span>
                {t(key)}
              </li>
            ))}
          </ul>
        </div>

        <MerchantConsolePreview />
      </div>
    </section>
  )
}
