import i18n from "i18next"
import { initReactI18next } from "react-i18next"
import {
  applyDocumentLanguage,
  DEFAULT_LANGUAGE,
  normalizeLanguage,
  rememberLanguage,
  resolveInitialLanguage,
  SUPPORTED_LANGUAGES,
} from "@/i18n/config"
import enAuth from "@/i18n/locales/en/auth.json"
import enCommon from "@/i18n/locales/en/common.json"
import enHome from "@/i18n/locales/en/home.json"
import enMerchant from "@/i18n/locales/en/merchant.json"
import zhCNAuth from "@/i18n/locales/zh-CN/auth.json"
import zhCNCommon from "@/i18n/locales/zh-CN/common.json"
import zhCNHome from "@/i18n/locales/zh-CN/home.json"
import zhCNMerchant from "@/i18n/locales/zh-CN/merchant.json"
import { BRAND } from "@/lib/brand"

export const NAMESPACES = ["common", "home", "auth", "merchant"] as const

export const resources = {
  "zh-CN": {
    common: zhCNCommon,
    home: zhCNHome,
    auth: zhCNAuth,
    merchant: zhCNMerchant,
  },
  en: {
    common: enCommon,
    home: enHome,
    auth: enAuth,
    merchant: enMerchant,
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
