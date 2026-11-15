import * as React from "react"
import { Search } from "lucide-react"
import { cn } from "@/shared/lib/utils"
import { Input } from "@/shared/ui/input"

type SearchInputProps = Omit<React.ComponentProps<typeof Input>, "type"> & {
  containerClassName?: string
}

export function SearchInput({
  className,
  containerClassName,
  ...props
}: SearchInputProps) {
  return (
    <div className={cn("relative w-full", containerClassName)}>
      <Search
        className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground"
        aria-hidden
      />
      <Input type="search" className={cn("pl-8", className)} {...props} />
    </div>
  )
}
