import { useEffect } from "react"
import { LoaderCircle, PackageOpen, Sparkles } from "lucide-react"
import { useInView } from "react-intersection-observer"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Badge } from "@/shared/ui/badge"
import { Button, buttonVariants } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { formatPriceCents, formatPriceRange } from "@/shared/lib/format"
import { describeFeedError } from "../model/errors"
import { useFeed } from "../model/queries"
import type { FeedItem } from "../types"

const IMAGE_ASPECT_CLASSES = [
  "aspect-[4/5]",
  "aspect-square",
  "aspect-[5/6]",
  "aspect-[4/3]",
  "aspect-[3/4]",
] as const

export function FeedPage({ token }: { token: string }) {
  const { t } = useTranslation("feed")
  const feedQuery = useFeed(token)
  const items = feedQuery.data?.pages.flatMap((page) => page.items) ?? []
  const total = feedQuery.data?.pages[0]?.total ?? 0
  const {
    error: feedError,
    fetchNextPage,
    hasNextPage,
    isFetching,
  } = feedQuery
  const { ref: loadMoreRef, inView } = useInView({
    rootMargin: "0px 0px 1200px 0px",
  })

  useEffect(() => {
    if (!inView || !hasNextPage || isFetching || feedError) {
      return
    }
    void fetchNextPage()
  }, [inView, hasNextPage, isFetching, feedError, fetchNextPage])

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("title")} description={t("description")} />

      {feedQuery.error && !feedQuery.data ? (
        <QueryError
          title={t("errors.loadTitle")}
          description={describeFeedError(feedQuery.error, t)}
          retryLabel={t("errors.retry")}
          onRetry={feedQuery.refetch}
        />
      ) : feedQuery.isPending ? (
        <FeedSkeleton />
      ) : items.length === 0 ? (
        <Card className="items-center gap-3 px-6 py-16 text-center">
          <PackageOpen
            className="size-10 text-muted-foreground"
            aria-hidden
          />
          <div className="space-y-1">
            <h2 className="font-heading text-lg font-medium">
              {t("empty.title")}
            </h2>
            <p className="max-w-md text-muted-foreground">
              {t("empty.description")}
            </p>
          </div>
        </Card>
      ) : (
        <>
          <div className="columns-2 gap-3 sm:columns-3 sm:gap-4 lg:columns-4">
            {items.map((item) => (
              <FeedCard key={item.id} item={item} />
            ))}
          </div>

          <div className="flex flex-col items-center gap-3 border-t pt-6">
            <p
              className="text-sm text-muted-foreground"
              aria-live="polite"
            >
              {t("progress", { loaded: items.length, total })}
            </p>
            {feedQuery.hasNextPage ? (
              <div
                ref={loadMoreRef}
                className="flex h-10 items-center justify-center text-sm text-muted-foreground"
                aria-live="polite"
              >
                {feedQuery.isFetchingNextPage ? (
                  <>
                    <LoaderCircle
                      className="mr-2 size-4 animate-spin"
                      aria-hidden
                    />
                    {t("loadingMore")}
                  </>
                ) : (
                  <span className="sr-only">{t("loadingMore")}</span>
                )}
              </div>
            ) : (
              <p className="text-xs text-muted-foreground">{t("end")}</p>
            )}
            {feedQuery.error ? (
              <div className="flex flex-col items-center gap-2">
                <p className="text-sm text-destructive">
                  {describeFeedError(feedQuery.error, t)}
                </p>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => void feedQuery.fetchNextPage()}
                >
                  {t("errors.retry")}
                </Button>
              </div>
            ) : null}
          </div>
        </>
      )}
    </div>
  )
}

export function FeedSignedOut() {
  const { t } = useTranslation("feed")

  return (
    <Card className="items-center gap-4 px-6 py-20 text-center">
      <div className="flex size-12 items-center justify-center rounded-full bg-muted">
        <Sparkles className="size-6 text-muted-foreground" aria-hidden />
      </div>
      <div className="space-y-1">
        <h1 className="font-heading text-2xl font-semibold text-balance">
          {t("signedOut.title")}
        </h1>
        <p className="max-w-md text-muted-foreground">
          {t("signedOut.description")}
        </p>
      </div>
      <Link to="/signin" className={buttonVariants()}>
        {t("signedOut.action")}
      </Link>
    </Card>
  )
}

function FeedCard({ item }: { item: FeedItem }) {
  const { t } = useTranslation("feed")
  const aspectClass =
    IMAGE_ASPECT_CLASSES[item.id % IMAGE_ASPECT_CLASSES.length]
  const price =
    formatPriceRange(item.min_price_cents, item.max_price_cents) ??
    formatPriceCents(item.min_price_cents)

  return (
    <Link
      to={`/products/${item.id}`}
      className="group mb-3 block break-inside-avoid rounded-xl focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50 sm:mb-4"
    >
      <Card className="flex w-full gap-0 p-0 transition-shadow group-hover:ring-foreground/25">
        <div
          role="img"
          aria-label={t("imagePlaceholder")}
          className={`w-full overflow-hidden ${aspectClass}`}
        >
          <Skeleton className="size-full rounded-none" />
        </div>
        <div className="flex flex-col gap-2 p-3">
          <div className="flex min-w-0 items-center gap-2">
            <Badge variant="secondary" className="shrink-0">
              {item.category_name ?? t("uncategorized")}
            </Badge>
            <span className="truncate text-xs text-muted-foreground">
              {item.shop_name}
            </span>
          </div>
          <h2 className="line-clamp-2 font-heading text-sm leading-snug font-medium">
            {item.name}
          </h2>
          {item.description ? (
            <p className="line-clamp-2 text-xs leading-relaxed text-muted-foreground">
              {item.description}
            </p>
          ) : null}
          <div className="mt-auto flex items-end justify-between gap-2 pt-1">
            <p className="font-heading text-base font-semibold tabular-nums">
              {price}
            </p>
            <p className="shrink-0 text-xs text-muted-foreground">
              {t("availableStock", { count: item.available_stock })}
            </p>
          </div>
        </div>
      </Card>
    </Link>
  )
}

function FeedSkeleton() {
  return (
    <div
      className="columns-2 gap-3 sm:columns-3 sm:gap-4 lg:columns-4"
      aria-hidden
    >
      {Array.from({ length: 8 }, (_, index) => (
        <Card
          key={index}
          className="mb-3 inline-flex w-full break-inside-avoid gap-0 p-0 sm:mb-4"
        >
          <Skeleton
            className={`w-full rounded-none ${
              IMAGE_ASPECT_CLASSES[index % IMAGE_ASPECT_CLASSES.length]
            }`}
          />
          <div className="space-y-2 p-3">
            <Skeleton className="h-3 w-20" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-3 w-2/3" />
          </div>
        </Card>
      ))}
    </div>
  )
}
