import en from "./en.json"
import zhCN from "./zh-CN.json"

export const mockPaymentResources = {
  "zh-CN": { mockpayment: zhCN },
  en: { mockpayment: en },
} as const
