import { Boxes, Package, Store } from "lucide-react"
import { useTranslation } from "react-i18next"
import { PREVIEW_PRODUCTS } from "./merchant-home-content"

export function MerchantConsolePreview() {
  const { t } = useTranslation()

  return (
    <div className="relative mx-auto w-full max-w-[31rem] lg:max-w-none">
      <div
        aria-hidden
        className="absolute -inset-6 rounded-[3rem] bg-gradient-to-br from-orange-500/20 via-rose-500/10 to-transparent blur-3xl"
      />

      <div className="relative overflow-hidden rounded-3xl border border-white/10 bg-neutral-950 text-white shadow-2xl shadow-neutral-950/25">
        <div className="flex items-center gap-1.5 border-b border-white/10 px-5 py-3.5">
          <span className="size-2 rounded-full bg-white/15" />
          <span className="size-2 rounded-full bg-white/15" />
          <span className="size-2 rounded-full bg-white/15" />
          <span className="ml-2 text-[11px] font-medium tracking-wide text-white/60">
            {t("home:merchant.preview.label")}
          </span>
          <span className="ml-auto inline-flex items-center gap-1.5 text-[11px] text-emerald-200">
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
              <p className="mt-1 text-xs text-white/55">
                {t("home:merchant.preview.shopMeta")}
              </p>
            </div>
            <span className="ml-auto rounded-full border border-emerald-400/20 bg-emerald-400/10 px-2.5 py-1 text-[10px] font-medium text-emerald-200">
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
              <span className="text-xs font-medium text-white/75">
                {t("home:merchant.preview.recent.title")}
              </span>
              <span className="text-[10px] text-white/55">
                {t("home:merchant.preview.recent.count")}
              </span>
            </div>
            <div className="space-y-2.5">
              {PREVIEW_PRODUCTS.map((product) => (
                <div key={product.key} className="flex items-center gap-2.5">
                  <span className="grid size-8 place-items-center rounded-lg bg-white/[0.06] text-white/55">
                    <Package className="size-3.5" aria-hidden />
                  </span>
                  <div className="min-w-0">
                    <p className="truncate text-xs font-medium text-white/85">
                      {t(product.nameKey)}
                    </p>
                    <p className="mt-0.5 truncate text-[10px] text-white/55">
                      {t(product.metaKey)}
                    </p>
                  </div>
                  <span className="ml-auto text-xs tabular-nums text-white/75">
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
              <p className="mt-0.5 text-[10px] text-orange-100/70">
                {t("home:merchant.preview.stockAlert.description")}
              </p>
            </div>
            <span className="ml-auto grid size-7 place-items-center rounded-full bg-orange-400/15 text-xs font-medium text-orange-100">
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
      <p className="truncate text-[10px] text-white/55">{label}</p>
      <p className="mt-1.5 text-lg font-semibold tracking-tight tabular-nums">
        {value}
      </p>
    </div>
  )
}
