import i18n from "@/shared/i18n/instance"
import { initReactI18next } from "react-i18next"
import {
  applyDocumentLanguage,
  DEFAULT_LANGUAGE,
  normalizeLanguage,
  rememberLanguage,
  resolveInitialLanguage,
  SUPPORTED_LANGUAGES,
} from "@/shared/i18n/config"
import enAuth from "./locales/en/auth.json"
import enCart from "./locales/en/cart.json"
import enCommon from "./locales/en/common.json"
import enConsole from "./locales/en/console.json"
import enFeed from "./locales/en/feed.json"
import enHome from "./locales/en/home.json"
import enOrder from "./locales/en/order.json"
import enPayment from "./locales/en/payment.json"
import zhCNAuth from "./locales/zh-CN/auth.json"
import zhCNCart from "./locales/zh-CN/cart.json"
import zhCNCommon from "./locales/zh-CN/common.json"
import zhCNConsole from "./locales/zh-CN/console.json"
import zhCNFeed from "./locales/zh-CN/feed.json"
import zhCNHome from "./locales/zh-CN/home.json"
import zhCNOrder from "./locales/zh-CN/order.json"
import zhCNPayment from "./locales/zh-CN/payment.json"
import { BRAND } from "@/shared/config/brand"
import { mockLogisticsResources } from "@/mock/logistics"
import { mockPaymentResources } from "@/mock/payment"

export const NAMESPACES = [
  "common",
  "home",
  "auth",
  "console",
  "cart",
  "order",
  "payment",
  "feed",
  "mockpayment",
  "mocklogistics",
] as const

export const resources = {
  "zh-CN": {
    common: zhCNCommon,
    home: zhCNHome,
    auth: zhCNAuth,
    console: zhCNConsole,
    cart: zhCNCart,
    order: zhCNOrder,
    payment: zhCNPayment,
    feed: zhCNFeed,
    ...mockPaymentResources["zh-CN"],
    ...mockLogisticsResources["zh-CN"],
  },
  en: {
    common: enCommon,
    home: enHome,
    auth: enAuth,
    console: enConsole,
    cart: enCart,
    order: enOrder,
    payment: enPayment,
    feed: enFeed,
    ...mockPaymentResources.en,
    ...mockLogisticsResources.en,
  },
}

const initialLanguage = resolveInitialLanguage()

void i18n.use(initReactI18next).init({
  resources,
  lng: initialLanguage,
  fallbackLng: DEFAULT_LANGUAGE,
  supportedLngs: SUPPORTED_LANGUAGES,
  defaultNS: "common",
  ns: NAMESPACES,
  interpolation: {
    escapeValue: false,
    defaultVariables: { brand: BRAND.name },
  },
  initAsync: false,
})

i18n.on("languageChanged", (language) => {
  const next = normalizeLanguage(language)
  rememberLanguage(next)
  applyDocumentLanguage(next)
})

applyDocumentLanguage(initialLanguage)

export default i18n
