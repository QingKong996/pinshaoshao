import { Package } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { buttonVariants } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

export function OrderSignedOut() {
  const { t } = useTranslation("order")
  useDocumentTitle(t("signedOut.documentTitle"))

  return (
    <Card className="items-center gap-3 px-6 py-16 text-center">
      <Package className="size-10 text-muted-foreground" aria-hidden />
      <div className="space-y-1">
        <h1 className="font-heading text-2xl font-semibold">
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
