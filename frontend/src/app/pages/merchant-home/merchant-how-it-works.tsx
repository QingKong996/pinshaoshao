import { ArrowRight } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { buttonVariants } from "@/shared/ui/button"
import { STEPS } from "./merchant-home-content"
import { MerchantSectionIntro } from "./merchant-section-intro"

export function MerchantHowItWorks() {
  const { t } = useTranslation()

  return (
    <section id="merchant-how-it-works" className="py-20 lg:py-24">
      <div className="mx-auto grid w-full max-w-6xl gap-10 px-4 sm:px-6 lg:grid-cols-[0.8fr_1.2fr] lg:gap-16">
        <div>
          <MerchantSectionIntro
            eyebrow={t("home:merchant.howItWorks.eyebrow")}
            title={t("home:merchant.howItWorks.title")}
            description={t("home:merchant.howItWorks.description")}
            descriptionClassName="max-w-md"
          />
          <Link
            to="/merchant/signup"
            className={buttonVariants({
              variant: "outline",
              size: "lg",
              className: "mt-7 h-10 rounded-xl",
            })}
          >
            {t("home:merchant.howItWorks.cta")}
            <ArrowRight aria-hidden />
          </Link>
        </div>

        <ol className="grid gap-3">
          {STEPS.map((step, index) => (
            <li
              key={step.key}
              className="flex gap-4 rounded-2xl border border-border/70 bg-card p-5 transition-colors hover:border-orange-500/25 sm:gap-5 sm:p-6"
            >
              <span className="grid size-9 shrink-0 place-items-center rounded-full border border-orange-500/20 bg-orange-500/10 text-xs font-semibold text-orange-600 dark:text-orange-400">
                {String(index + 1).padStart(2, "0")}
              </span>
              <div>
                <h3 className="font-heading text-base font-semibold">
                  {t(step.titleKey)}
                </h3>
                <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                  {t(step.descriptionKey)}
                </p>
              </div>
            </li>
          ))}
        </ol>
      </div>
    </section>
  )
}
