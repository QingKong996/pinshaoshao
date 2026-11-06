import { ChevronDown, Languages } from "lucide-react"
import { useTranslation } from "react-i18next"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  normalizeLanguage,
  SUPPORTED_LANGUAGES,
  type Language,
} from "@/i18n/config"
import { cn } from "@/lib/utils"

type LanguageSwitcherProps = {
  className?: string
}

export function LanguageSwitcher({ className }: LanguageSwitcherProps) {
  const { t, i18n } = useTranslation()
  const activeLanguage = normalizeLanguage(
    i18n.resolvedLanguage ?? i18n.language,
  )

  function changeLanguage(language: Language) {
    if (language === activeLanguage) {
      return
    }
    void i18n.changeLanguage(language)
  }

  return (
    <div className={cn("inline-flex items-center", className)}>
      <DropdownMenu>
        <DropdownMenuTrigger
          aria-label={t("common:language.switchLabel")}
          className="inline-flex h-8 cursor-pointer items-center gap-1.5 rounded-lg px-2 text-sm font-medium whitespace-nowrap text-muted-foreground transition-colors outline-none hover:bg-muted hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
        >
          <Languages className="size-4" aria-hidden />
          <span>{t(`common:language.options.${activeLanguage}`)}</span>
          <ChevronDown className="size-3.5 opacity-60" aria-hidden />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-36">
          <DropdownMenuRadioGroup
            value={activeLanguage}
            onValueChange={(value) => changeLanguage(normalizeLanguage(value))}
          >
            {SUPPORTED_LANGUAGES.map((language) => (
              <DropdownMenuRadioItem
                key={language}
                value={language}
                lang={language}
                closeOnClick
              >
                {t(`common:language.options.${language}`)}
              </DropdownMenuRadioItem>
            ))}
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
