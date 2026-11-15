import { cn } from "@/shared/lib/utils"

export function MerchantSectionIntro({
  eyebrow,
  title,
  description,
  className,
  descriptionClassName,
}: {
  eyebrow: string
  title: string
  description: string
  className?: string
  descriptionClassName?: string
}) {
  return (
    <div className={cn("max-w-2xl", className)}>
      <p className="text-sm font-medium text-orange-600 dark:text-orange-400">
        {eyebrow}
      </p>
      <h2 className="mt-3 font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
        {title}
      </h2>
      <p
        className={cn(
          "mt-4 text-base leading-7 text-muted-foreground",
          descriptionClassName,
        )}
      >
        {description}
      </p>
    </div>
  )
}
