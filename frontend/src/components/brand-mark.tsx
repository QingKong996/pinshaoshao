import { Link } from "react-router"
import { BRAND } from "@/config/brand"
import { cn } from "@/lib/utils"

type BrandMarkProps = {
  className?: string
  inverted?: boolean
}

const TAG_PATH =
  "M12.646 2.646A2 2 0 0 0 11.232 2H4a2 2 0 0 0-2 2v7.232a2 2 0 0 0 .586 1.414l8.704 8.704a2.426 2.426 0 0 0 3.42 0l6.58-6.58a2.426 2.426 0 0 0 0-3.42zM6.15 7.5a1.35 1.35 0 1 0 2.7 0 1.35 1.35 0 1 0-2.7 0z"

const SPARKLE_PATH =
  "M19.2 0.8Q19.92 4.58 23.7 5.3Q19.92 6.02 19.2 9.8Q18.48 6.02 14.7 5.3Q18.48 4.58 19.2 0.8Z"

function BrandGlyph() {
  return (
    <svg
      viewBox="0 0 24 24"
      aria-hidden="true"
      focusable="false"
      className="size-[22px]"
    >
      <g transform="translate(12 12) rotate(-8) scale(.82) translate(-12 -12)">
        <path fillRule="evenodd" fill="currentColor" d={TAG_PATH} />
      </g>
      <path fill="currentColor" d={SPARKLE_PATH} />
    </svg>
  )
}

export function BrandMark({ className, inverted = false }: BrandMarkProps) {
  return (
    <Link
      to="/"
      aria-label={BRAND.name}
      className={cn(
        "group inline-flex items-center gap-2.5 rounded-lg outline-none focus-visible:ring-3 focus-visible:ring-ring/50",
        className,
      )}
    >
      <span className="grid size-9 place-items-center rounded-xl bg-gradient-to-br from-orange-500 via-rose-500 to-rose-600 text-white shadow-sm shadow-rose-500/20 transition-transform duration-200 group-hover:-rotate-3 group-hover:scale-105">
        <BrandGlyph />
      </span>
      <span className="flex flex-col leading-none">
        <span
          className={cn(
            "text-base font-semibold tracking-tight",
            inverted ? "text-white" : "text-foreground",
          )}
        >
          {BRAND.name}
        </span>
        <span
          className={cn(
            "mt-1 text-[10px] font-medium tracking-[0.2em]",
            inverted ? "text-white/50" : "text-muted-foreground",
          )}
        >
          {BRAND.wordmark}
        </span>
      </span>
    </Link>
  )
}
