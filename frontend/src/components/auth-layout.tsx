import type { ReactNode } from "react"
import {
  Package,
  ShieldCheck,
  Store,
  TrendingUp,
  Truck,
  Users,
  X,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { BrandMark } from "@/components/brand-mark"
import { LanguageSwitcher } from "@/components/language-switcher"
import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

type AuthLayoutProps = {
  title: string
  description: string
  footer: ReactNode
  children: ReactNode
  variant?: "user" | "merchant"
}

const USER_HIGHLIGHTS = [
  { icon: Users, labelKey: "auth:layout.highlights.groupPrice" },
  { icon: ShieldCheck, labelKey: "auth:layout.highlights.authentic" },
  { icon: Truck, labelKey: "auth:layout.highlights.service" },
] as const

const MERCHANT_HIGHLIGHTS = [
  { icon: Store, labelKey: "auth:merchant.layout.highlights.storefront" },
  { icon: Package, labelKey: "auth:merchant.layout.highlights.products" },
  { icon: TrendingUp, labelKey: "auth:merchant.layout.highlights.orders" },
] as const

const USER_PANEL = {
  titleLead: "auth:layout.titleLead",
  titleRest: "auth:layout.titleRest",
  description: "auth:layout.description",
} as const

const MERCHANT_PANEL = {
  titleLead: "auth:merchant.layout.titleLead",
  titleRest: "auth:merchant.layout.titleRest",
  description: "auth:merchant.layout.description",
} as const

export function AuthLayout({
  title,
  description,
  footer,
  children,
  variant = "user",
}: AuthLayoutProps) {
  const { t } = useTranslation()
  const panel = variant === "merchant" ? MERCHANT_PANEL : USER_PANEL
  const highlights =
    variant === "merchant" ? MERCHANT_HIGHLIGHTS : USER_HIGHLIGHTS

  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      <aside className="relative hidden flex-col justify-between overflow-hidden bg-neutral-950 p-10 text-white lg:flex xl:p-14">
        <div aria-hidden className="pointer-events-none absolute inset-0">
          <div className="absolute -top-32 -left-24 size-96 rounded-full bg-orange-500/25 blur-3xl" />
          <div className="absolute top-1/3 -right-24 size-80 rounded-full bg-rose-500/20 blur-3xl" />
          <div className="absolute inset-0 bg-[linear-gradient(to_right,rgba(255,255,255,0.05)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.05)_1px,transparent_1px)] bg-[size:56px_56px] [mask-image:radial-gradient(ellipse_at_center,black_35%,transparent_75%)]" />
        </div>

        <div className="relative">
          <BrandMark inverted />
        </div>

        <div className="relative max-w-md">
          <h2 className="text-3xl font-semibold tracking-tight text-balance xl:text-4xl">
            {t(panel.titleLead)}
            <br />
            {t(panel.titleRest)}
          </h2>
          <p className="mt-4 text-sm leading-relaxed text-white/60">
            {t(panel.description)}
          </p>
          <ul className="mt-8 space-y-3">
            {highlights.map(({ icon: Icon, labelKey }) => (
              <li
                key={labelKey}
                className="flex items-center gap-3 text-sm text-white/75"
              >
                <span className="grid size-8 place-items-center rounded-lg border border-white/10 bg-white/5">
                  <Icon className="size-4 text-orange-300" />
                </span>
                {t(labelKey)}
              </li>
            ))}
          </ul>
        </div>

        <p className="relative text-xs text-white/35">
          {t("auth:layout.disclaimer")}
        </p>
      </aside>

      <main className="flex min-h-svh flex-col bg-background">
        <div className="flex items-center justify-between px-5 py-5 sm:px-8">
          <div className="lg:hidden">
            <BrandMark />
          </div>
          <div className="ml-auto flex items-center gap-3">
            <LanguageSwitcher />
            <Link
              to="/"
              aria-label={t("auth:layout.backHome")}
              className={cn(
                buttonVariants({ variant: "outline", size: "icon" }),
              )}
            >
              <X aria-hidden />
            </Link>
          </div>
        </div>

        <div className="flex flex-1 items-center justify-center px-5 pt-2 pb-16 sm:px-8">
          <div className="w-full max-w-sm">
            <div className="mb-8">
              <h1 className="font-heading text-2xl font-semibold tracking-tight text-balance sm:text-3xl">
                {title}
              </h1>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {description}
              </p>
            </div>
            {children}
            <div className="mt-6 text-center text-sm text-muted-foreground">
              {footer}
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
