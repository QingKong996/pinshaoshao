import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Check, Copy } from "lucide-react"
import { Button } from "@/shared/ui/button"

export function CopyButton({
  value,
  label,
}: {
  value: string
  label?: string
}) {
  const { t } = useTranslation("mockpayment")
  const [copied, setCopied] = useState(false)
  const timeoutRef = useRef<number | undefined>(undefined)

  useEffect(() => {
    return () => window.clearTimeout(timeoutRef.current)
  }, [])

  async function copy() {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      window.clearTimeout(timeoutRef.current)
      timeoutRef.current = window.setTimeout(() => setCopied(false), 2_000)
    } catch {
      setCopied(false)
    }
  }

  return (
    <Button type="button" variant="outline" size="sm" onClick={() => void copy()}>
      {copied ? <Check aria-hidden /> : <Copy aria-hidden />}
      {copied ? t("actions.copied") : (label ?? t("actions.copy"))}
    </Button>
  )
}
