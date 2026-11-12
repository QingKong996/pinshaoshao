import { LogOut, User as UserIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAuth, type UserSession } from "@/auth/user/context"
import { cn } from "@/lib/utils"

type UserMenuProps = {
  session: UserSession
  className?: string
}

export function UserMenu({ session, className }: UserMenuProps) {
  const { signOut } = useAuth()
  const { t } = useTranslation()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        openOnHover
        closeDelay={120}
        aria-label={t("common:auth.accountMenu")}
        className={cn(
          "inline-flex min-w-0 cursor-pointer items-center gap-2 rounded-lg p-1 text-sm font-medium outline-none transition-colors hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/50 sm:pr-2.5",
          className,
        )}
      >
        <Avatar className="size-7">
          <AvatarFallback>
            <UserIcon className="size-4" aria-hidden />
          </AvatarFallback>
        </Avatar>
        <span className="hidden max-w-32 truncate sm:inline">
          {session.username}
        </span>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" sideOffset={8} className="w-44">
        <DropdownMenuItem onClick={() => void signOut()}>
          <LogOut aria-hidden />
          {t("common:auth.signOut")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
