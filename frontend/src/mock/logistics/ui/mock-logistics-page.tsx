import { ArrowRight, PackageCheck, Store, Webhook } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { CreateShipmentCard } from "./create-shipment-card"
import { MockLogisticsShell } from "./mock-logistics-shell"
import { ShipmentLookupCard } from "./shipment-lookup-card"

export function MockLogisticsPage() {
  const { t } = useTranslation("mocklogistics")
  useDocumentTitle(t("documentTitle.console"))

  return (
    <MockLogisticsShell>
      <div className="flex flex-col gap-6">
        <PageHeader
          title={t("console.title")}
          description={t("console.description")}
          actions={
            <Badge variant="secondary" className="font-mono">
              API v1
            </Badge>
          }
        />

        <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1.45fr)_minmax(18rem,0.55fr)]">
          <CreateShipmentCard />
          <div className="flex flex-col gap-6">
            <ShipmentLookupCard />
            <WorkflowGuide />
          </div>
        </div>
      </div>
    </MockLogisticsShell>
  )
}

function WorkflowGuide() {
  const { t } = useTranslation("mocklogistics")
  const steps = [
    {
      icon: PackageCheck,
      title: t("console.guide.create.title"),
      description: t("console.guide.create.description"),
    },
    {
      icon: Store,
      title: t("console.guide.pickup.title"),
      description: t("console.guide.pickup.description"),
    },
    {
      icon: Webhook,
      title: t("console.guide.advance.title"),
      description: t("console.guide.advance.description"),
    },
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("console.guide.title")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-5">
        <ol className="grid gap-4">
          {steps.map((step, index) => {
            const Icon = step.icon
            return (
              <li key={step.title} className="flex gap-3">
                <span className="grid size-8 shrink-0 place-items-center rounded-lg bg-muted text-muted-foreground">
                  <Icon className="size-4" aria-hidden />
                </span>
                <div className="space-y-1">
                  <p className="text-sm font-medium">
                    {index + 1}. {step.title}
                  </p>
                  <p className="text-sm leading-relaxed text-muted-foreground">
                    {step.description}
                  </p>
                </div>
              </li>
            )
          })}
        </ol>
        <div className="flex gap-2 rounded-lg border border-dashed p-3 text-xs leading-relaxed text-muted-foreground">
          <ArrowRight className="mt-0.5 size-3.5 shrink-0" aria-hidden />
          <p>{t("console.guide.isolation")}</p>
        </div>
      </CardContent>
    </Card>
  )
}
