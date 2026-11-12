import { useState, type ComponentProps } from "react"
import { Eye, EyeOff } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Input } from "@/shared/ui/input"
import { cn } from "@/shared/lib/utils"

type PasswordInputProps = Omit<ComponentProps<"input">, "type">

export function PasswordInput({
  className,
  disabled,
  ref,
  ...props
}: PasswordInputProps) {
  const { t } = useTranslation()
  const [visible, setVisible] = useState(false)

  return (
    <div className="relative">
      <Input
        ref={ref}
        type={visible ? "text" : "password"}
        disabled={disabled}
        className={cn("pr-11", className)}
        {...props}
      />
      <button
        type="button"
        onClick={() => setVisible((current) => !current)}
        disabled={disabled}
        aria-label={
          visible ? t("auth:password.hide") : t("auth:password.show")
        }
        className="absolute top-1/2 right-1.5 grid size-8 -translate-y-1/2 cursor-pointer place-items-center rounded-md text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50"
      >
        {visible ? (
          <EyeOff className="size-4" aria-hidden />
        ) : (
          <Eye className="size-4" aria-hidden />
        )}
      </button>
    </div>
  )
}
