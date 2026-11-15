import { ArrowRight, Check } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Badge } from "@/shared/ui/badge"
import { buttonVariants } from "@/shared/ui/button"
import { ASSURANCES } from "./merchant-home-content"

export function MerchantClosingCta() {
  const { t } = useTranslation()

  return (
    <section className="mx-auto w-full max-w-6xl px-4 pb-20 sm:px-6 lg:pb-24">
      <div className="relative overflow-hidden rounded-3xl bg-neutral-950 px-6 py-14 text-white sm:px-10 lg:px-14 lg:py-16">
        <div aria-hidden className="pointer-events-none absolute inset-0">
          <div className="absolute -top-28 -left-16 size-80 rounded-full bg-orange-500/20 blur-3xl" />
          <div className="absolute -right-16 bottom-0 size-72 rounded-full bg-rose-500/15 blur-3xl" />
          <div className="absolute inset-0 bg-[linear-gradient(to_right,rgba(255,255,255,0.045)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.045)_1px,transparent_1px)] bg-[size:56px_56px] [mask-image:radial-gradient(ellipse_at_center,black_25%,transparent_75%)]" />
        </div>

        <div className="relative max-w-2xl">
          <Badge
            variant="outline"
            className="border-white/15 bg-white/5 text-white/75"
          >
            {t("home:merchant.closingCta.badge")}
          </Badge>
          <h2 className="mt-5 font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
            {t("home:merchant.closingCta.title.lead")}
            <span className="mt-1 block text-orange-300">
              {t("home:merchant.closingCta.title.highlight")}
            </span>
          </h2>
          <p className="mt-4 max-w-xl text-sm leading-7 text-white/70 sm:text-base">
            {t("home:merchant.closingCta.description")}
          </p>

          <div className="mt-7 flex flex-col gap-3 sm:flex-row">
            <Link
              to="/merchant/signup"
              className={buttonVariants({
                variant: "secondary",
                size: "lg",
                className: "h-11 rounded-xl bg-white px-5 hover:bg-white/90",
              })}
            >
              {t("home:merchant.closingCta.actions.signUp")}
              <ArrowRight aria-hidden />
            </Link>
            <Link
              to="/merchant/signin"
              className={buttonVariants({
                variant: "outline",
                size: "lg",
                className:
                  "h-11 rounded-xl border-white/15 bg-white/5 px-5 text-white hover:bg-white/10 hover:text-white",
              })}
            >
              {t("home:merchant.closingCta.actions.signIn")}
            </Link>
          </div>

          <ul className="mt-7 flex flex-wrap gap-x-5 gap-y-2.5">
            {ASSURANCES.map((key) => (
              <li
                key={key}
                className="flex items-center gap-2 text-xs text-white/65"
              >
                <Check className="size-3.5 text-orange-300" aria-hidden />
                {t(key)}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  )
}
