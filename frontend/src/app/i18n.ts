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
import enCommon from "./locales/en/common.json"
import enConsole from "./locales/en/console.json"
import enHome from "./locales/en/home.json"
import zhCNAuth from "./locales/zh-CN/auth.json"
import zhCNCommon from "./locales/zh-CN/common.json"
import zhCNConsole from "./locales/zh-CN/console.json"
import zhCNHome from "./locales/zh-CN/home.json"
import { BRAND } from "@/shared/config/brand"

export const NAMESPACES = ["common", "home", "auth", "console"] as const

export const resources = {
  "zh-CN": {
    common: zhCNCommon,
    home: zhCNHome,
    auth: zhCNAuth,
    console: zhCNConsole,
  },
  en: {
    common: enCommon,
    home: enHome,
    auth: enAuth,
    console: enConsole,
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
