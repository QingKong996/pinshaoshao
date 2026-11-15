import {
  ArrowRight,
  Boxes,
  Check,
  LayoutDashboard,
  Package,
  Store,
  Users,
  Warehouse,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { MerchantSiteFooter } from "../layouts/merchant-site-footer"
import { MerchantSiteHeader } from "../layouts/merchant-site-header"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { Badge } from "@/shared/ui/badge"
import { buttonVariants } from "@/shared/ui/button"
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"

const FEATURES = [
  {
    key: "reach",
    icon: Users,
    titleKey: "home:merchant.features.items.reach.title",
    descriptionKey: "home:merchant.features.items.reach.description",
  },
  {
    key: "catalog",
    icon: Package,
    titleKey: "home:merchant.features.items.catalog.title",
    descriptionKey: "home:merchant.features.items.catalog.description",
  },
  {
    key: "inventory",
    icon: Warehouse,
    titleKey: "home:merchant.features.items.inventory.title",
    descriptionKey: "home:merchant.features.items.inventory.description",
  },
  {
    key: "overview",
    icon: LayoutDashboard,
    titleKey: "home:merchant.features.items.overview.title",
    descriptionKey: "home:merchant.features.items.overview.description",
  },
] as const

const STEPS = [
  {
    key: "register",
    titleKey: "home:merchant.howItWorks.steps.register.title",
    descriptionKey: "home:merchant.howItWorks.steps.register.description",
  },
  {
    key: "publish",
    titleKey: "home:merchant.howItWorks.steps.publish.title",
    descriptionKey: "home:merchant.howItWorks.steps.publish.description",
  },
  {
    key: "manage",
    titleKey: "home:merchant.howItWorks.steps.manage.title",
    descriptionKey: "home:merchant.howItWorks.steps.manage.description",
  },
] as const

const PREVIEW_PRODUCTS = [
  {
    key: "coffee",
    nameKey: "home:merchant.preview.recent.items.coffee.name",
    metaKey: "home:merchant.preview.recent.items.coffee.meta",
    price: "¥68",
  },
  {
    key: "bag",
    nameKey: "home:merchant.preview.recent.items.bag.name",
    metaKey: "home:merchant.preview.recent.items.bag.meta",
    price: "¥129",
  },
] as const

const TRUST_POINTS = [
  "home:merchant.hero.trustPoints.free",
  "home:merchant.hero.trustPoints.management",
  "home:merchant.hero.trustPoints.insight",
] as const

const ASSURANCES = [
  "home:merchant.closingCta.assurances.free",
  "home:merchant.closingCta.assurances.setup",
  "home:merchant.closingCta.assurances.mobile",
] as const

export default function MerchantHomePage() {
  const { t } = useTranslation()
  useDocumentTitle(t("home:merchant.documentTitle"))

  return (
    <div className="flex min-h-svh flex-col">
      <MerchantSiteHeader />
      <main className="flex-1">
        <MerchantHero />
        <MerchantFeatures />
        <MerchantHowItWorks />
        <MerchantClosingCta />
      </main>
      <MerchantSiteFooter />
    </div>
  )
}

function MerchantHero() {
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

function MerchantConsolePreview() {
  const { t } = useTranslation()

  return (
    <div className="relative mx-auto w-full max-w-[31rem] lg:max-w-none">
      <div
        aria-hidden
        className="absolute -inset-6 rounded-[3rem] bg-gradient-to-br from-orange-500/20 via-rose-500/10 to-transparent blur-3xl"
      />

      <div className="relative overflow-hidden rounded-[1.75rem] border border-white/10 bg-neutral-950 text-white shadow-2xl shadow-neutral-950/25">
        <div className="flex items-center gap-1.5 border-b border-white/10 px-5 py-3.5">
          <span className="size-2 rounded-full bg-white/15" />
          <span className="size-2 rounded-full bg-white/15" />
          <span className="size-2 rounded-full bg-white/15" />
          <span className="ml-2 text-[11px] font-medium tracking-wide text-white/40">
            {t("home:merchant.preview.label")}
          </span>
          <span className="ml-auto inline-flex items-center gap-1.5 text-[11px] text-emerald-300">
            <span className="size-1.5 rounded-full bg-emerald-400" />
            {t("home:merchant.preview.active")}
          </span>
        </div>

        <div className="space-y-4 p-5 sm:p-6">
          <div className="flex items-center gap-3">
            <span className="grid size-11 place-items-center rounded-xl bg-gradient-to-br from-orange-500 to-rose-600 shadow-lg shadow-rose-950/30">
              <Store className="size-5" aria-hidden />
            </span>
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">
                {t("home:merchant.preview.shopName")}
              </p>
              <p className="mt-1 text-xs text-white/40">
                {t("home:merchant.preview.shopMeta")}
              </p>
            </div>
            <span className="ml-auto rounded-full border border-emerald-400/20 bg-emerald-400/10 px-2.5 py-1 text-[10px] font-medium text-emerald-300">
              {t("home:merchant.preview.status")}
            </span>
          </div>

          <div className="grid grid-cols-3 gap-2">
            <PreviewStat
              label={t("home:merchant.preview.stats.products")}
              value="24"
            />
            <PreviewStat
              label={t("home:merchant.preview.stats.skus")}
              value="86"
            />
            <PreviewStat
              label={t("home:merchant.preview.stats.stock")}
              value="1,284"
            />
          </div>

          <div className="rounded-xl border border-white/10 bg-white/[0.035] p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <span className="text-xs font-medium text-white/70">
                {t("home:merchant.preview.recent.title")}
              </span>
              <span className="text-[10px] text-white/35">
                {t("home:merchant.preview.recent.count")}
              </span>
            </div>
            <div className="space-y-2.5">
              {PREVIEW_PRODUCTS.map((product) => (
                <div key={product.key} className="flex items-center gap-2.5">
                  <span className="grid size-8 place-items-center rounded-lg bg-white/[0.06] text-white/45">
                    <Package className="size-3.5" aria-hidden />
                  </span>
                  <div className="min-w-0">
                    <p className="truncate text-xs font-medium text-white/80">
                      {t(product.nameKey)}
                    </p>
                    <p className="mt-0.5 truncate text-[10px] text-white/35">
                      {t(product.metaKey)}
                    </p>
                  </div>
                  <span className="ml-auto text-xs tabular-nums text-white/65">
                    {product.price}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div className="flex items-center gap-3 rounded-xl border border-orange-400/20 bg-orange-400/[0.07] p-3.5">
            <span className="grid size-9 place-items-center rounded-lg bg-orange-400/10 text-orange-300">
              <Boxes className="size-4" aria-hidden />
            </span>
            <div>
              <p className="text-xs font-medium text-orange-100">
                {t("home:merchant.preview.stockAlert.title")}
              </p>
              <p className="mt-0.5 text-[10px] text-orange-200/55">
                {t("home:merchant.preview.stockAlert.description")}
              </p>
            </div>
            <span className="ml-auto grid size-7 place-items-center rounded-full bg-orange-400/15 text-xs font-medium text-orange-200">
              3
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}

function PreviewStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-white/10 bg-white/[0.04] p-3">
      <p className="truncate text-[10px] text-white/35">{label}</p>
      <p className="mt-1.5 text-lg font-semibold tracking-tight tabular-nums">
        {value}
      </p>
    </div>
  )
}

function MerchantFeatures() {
  const { t } = useTranslation()

  return (
    <section
      id="merchant-features"
      className="border-y border-border/60 bg-muted/20 py-20 lg:py-24"
    >
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <div className="max-w-2xl">
          <p className="text-sm font-medium text-orange-600 dark:text-orange-400">
            {t("home:merchant.features.eyebrow")}
          </p>
          <h2 className="mt-3 font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
            {t("home:merchant.features.title")}
          </h2>
          <p className="mt-4 text-base leading-7 text-muted-foreground">
            {t("home:merchant.features.description")}
          </p>
        </div>

        <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {FEATURES.map(({ key, icon: Icon, titleKey, descriptionKey }) => (
            <Card
              key={key}
              className="border border-border/60 bg-background/80 shadow-none transition-colors hover:border-orange-500/25"
            >
              <CardHeader>
                <span className="mb-2 grid size-10 place-items-center rounded-xl bg-orange-500/10 text-orange-600 ring-1 ring-orange-500/10 dark:text-orange-400">
                  <Icon className="size-4" aria-hidden />
                </span>
                <CardTitle>{t(titleKey)}</CardTitle>
                <CardDescription className="leading-relaxed">
                  {t(descriptionKey)}
                </CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>
      </div>
    </section>
  )
}

function MerchantHowItWorks() {
  const { t } = useTranslation()

  return (
    <section id="merchant-how-it-works" className="py-20 lg:py-24">
      <div className="mx-auto grid w-full max-w-6xl gap-10 px-4 sm:px-6 lg:grid-cols-[0.8fr_1.2fr] lg:gap-16">
        <div>
          <p className="text-sm font-medium text-orange-600 dark:text-orange-400">
            {t("home:merchant.howItWorks.eyebrow")}
          </p>
          <h2 className="mt-3 font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
            {t("home:merchant.howItWorks.title")}
          </h2>
          <p className="mt-4 max-w-md text-base leading-7 text-muted-foreground">
            {t("home:merchant.howItWorks.description")}
          </p>
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
              className="group flex gap-4 rounded-2xl border border-border/70 bg-card p-5 transition-colors hover:border-orange-500/25 sm:gap-5 sm:p-6"
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

function MerchantClosingCta() {
  const { t } = useTranslation()

  return (
    <section className="mx-auto w-full max-w-6xl px-4 pb-20 sm:px-6 lg:pb-24">
      <div className="relative overflow-hidden rounded-[2rem] bg-neutral-950 px-6 py-14 text-white sm:px-10 lg:px-14 lg:py-16">
        <div aria-hidden className="pointer-events-none absolute inset-0">
          <div className="absolute -top-28 -left-16 size-80 rounded-full bg-orange-500/20 blur-3xl" />
          <div className="absolute -right-16 bottom-0 size-72 rounded-full bg-rose-500/15 blur-3xl" />
          <div className="absolute inset-0 bg-[linear-gradient(to_right,rgba(255,255,255,0.045)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.045)_1px,transparent_1px)] bg-[size:56px_56px] [mask-image:radial-gradient(ellipse_at_center,black_25%,transparent_75%)]" />
        </div>

        <div className="relative max-w-2xl">
          <Badge
            variant="outline"
            className="border-white/15 bg-white/5 text-white/70"
          >
            {t("home:merchant.closingCta.badge")}
          </Badge>
          <h2 className="mt-5 font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
            {t("home:merchant.closingCta.title.lead")}
            <span className="mt-1 block text-orange-300">
              {t("home:merchant.closingCta.title.highlight")}
            </span>
          </h2>
          <p className="mt-4 max-w-xl text-sm leading-7 text-white/55 sm:text-base">
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
                className="flex items-center gap-2 text-xs text-white/45"
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
