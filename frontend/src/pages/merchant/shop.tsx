import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import {
  AlertCircle,
  CheckCircle2,
  LoaderCircle,
} from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { PageHeader } from "@/components/merchant/page-header"
import { QueryError } from "@/components/merchant/query-error"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { formatDateTime } from "@/lib/format"
import { getShopProfile, updateShopProfile } from "@/lib/merchant-api"
import { describeMerchantError } from "@/lib/merchant-errors"
import {
  merchantShopProfileSchema,
  translateMerchantValidationMessage,
  type MerchantShopProfileValues,
} from "@/lib/merchant-schemas"
import { useDocumentTitle } from "@/lib/use-document-title"
import { useMerchantSession } from "@/lib/use-merchant-session"
import { useRequest } from "@/lib/use-request"

const INITIAL_VALUES: MerchantShopProfileValues = {
  name: "",
  description: "",
}

export default function ShopProfilePage() {
  const { t } = useTranslation("merchant")
  useDocumentTitle(t("documentTitle.shop"))

  const { token } = useMerchantSession()
  const profile = useRequest(() => getShopProfile(token), [token])
  const [saved, setSaved] = useState(false)
  const [saveError, setSaveError] = useState<unknown>(null)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<MerchantShopProfileValues>({
    resolver: zodResolver(merchantShopProfileSchema),
    defaultValues: INITIAL_VALUES,
  })

  useEffect(() => {
    if (!profile.data) {
      return
    }
    reset({
      name: profile.data.name || profile.data.username,
      description: profile.data.description,
    })
  }, [profile.data, reset])

  function clearFeedback() {
    setSaved(false)
    setSaveError(null)
  }

  async function onSubmit(values: MerchantShopProfileValues) {
    setSaved(false)
    setSaveError(null)
    try {
      const updated = await updateShopProfile(token, values)
      reset({ name: updated.name, description: updated.description })
      setSaved(true)
    } catch (error) {
      setSaveError(error)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("shop.title")} description={t("shop.description")} />

      {profile.error ? (
        <QueryError error={profile.error} onRetry={profile.reload} />
      ) : null}

      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>{t("shop.form.profileTitle")}</CardTitle>
          </CardHeader>
          <CardContent>
            {profile.loading && !profile.data ? (
              <div className="flex flex-col gap-4">
                <Skeleton className="h-9 w-full" />
                <Skeleton className="h-24 w-full" />
                <Skeleton className="h-8 w-24" />
              </div>
            ) : profile.data ? (
              <form
                noValidate
                aria-busy={isSubmitting}
                onSubmit={handleSubmit(onSubmit)}
                className="flex flex-col gap-5"
              >
                {saveError ? (
                  <Alert variant="destructive">
                    <AlertCircle aria-hidden />
                    <AlertDescription>
                      {describeMerchantError(saveError, t)}
                    </AlertDescription>
                  </Alert>
                ) : null}
                {saved ? (
                  <Alert>
                    <CheckCircle2 aria-hidden />
                    <AlertDescription>{t("shop.saved")}</AlertDescription>
                  </Alert>
                ) : null}

                <FieldGroup>
                  <Field data-invalid={errors.name ? true : undefined}>
                    <FieldLabel htmlFor="shop-name">
                      {t("shop.form.nameLabel")}
                    </FieldLabel>
                    <Input
                      id="shop-name"
                      autoComplete="organization"
                      required
                      aria-invalid={errors.name ? true : undefined}
                      aria-describedby={
                        errors.name
                          ? "shop-name-hint shop-name-error"
                          : "shop-name-hint"
                      }
                      className="h-10"
                      {...register("name", { onChange: clearFeedback })}
                    />
                    <FieldDescription id="shop-name-hint">
                      {t("shop.form.nameHint")}
                    </FieldDescription>
                    {errors.name ? (
                      <FieldError id="shop-name-error">
                        {translateMerchantValidationMessage(
                          errors.name.message,
                          t,
                        )}
                      </FieldError>
                    ) : null}
                  </Field>

                  <Field data-invalid={errors.description ? true : undefined}>
                    <FieldLabel htmlFor="shop-description">
                      {t("shop.form.descriptionLabel")}
                    </FieldLabel>
                    <Textarea
                      id="shop-description"
                      rows={4}
                      aria-invalid={errors.description ? true : undefined}
                      aria-describedby={
                        errors.description
                          ? "shop-description-hint shop-description-error"
                          : "shop-description-hint"
                      }
                      {...register("description", {
                        onChange: clearFeedback,
                      })}
                    />
                    <FieldDescription id="shop-description-hint">
                      {t("shop.form.descriptionHint")}
                    </FieldDescription>
                    {errors.description ? (
                      <FieldError id="shop-description-error">
                        {translateMerchantValidationMessage(
                          errors.description.message,
                          t,
                        )}
                      </FieldError>
                    ) : null}
                  </Field>
                </FieldGroup>

                <div>
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="gap-2"
                  >
                    {isSubmitting ? (
                      <LoaderCircle className="animate-spin" aria-hidden />
                    ) : null}
                    {isSubmitting ? t("shop.saving") : t("shop.save")}
                  </Button>
                </div>
              </form>
            ) : null}
          </CardContent>
        </Card>

        {profile.data || !profile.error ? (
          <Card>
            <CardHeader>
              <CardTitle>{t("shop.form.accountTitle")}</CardTitle>
              <CardDescription>{t("shop.form.usernameHint")}</CardDescription>
            </CardHeader>
            <CardContent>
              {profile.data ? (
                <dl className="flex flex-col gap-3 text-sm">
                  <div className="flex items-center justify-between gap-3">
                    <dt className="text-muted-foreground">
                      {t("shop.form.usernameLabel")}
                    </dt>
                    <dd className="truncate font-medium">
                      {profile.data.username}
                    </dd>
                  </div>
                  <div className="flex items-center justify-between gap-3">
                    <dt className="text-muted-foreground">
                      {t("shop.form.emailLabel")}
                    </dt>
                    <dd className="truncate">{profile.data.email}</dd>
                  </div>
                  <div className="flex items-center justify-between gap-3">
                    <dt className="text-muted-foreground">
                      {t("shop.form.statusLabel")}
                    </dt>
                    <dd>
                      <Badge
                        variant={
                          profile.data.status === "active"
                            ? "secondary"
                            : "destructive"
                        }
                      >
                        {t(`status.${profile.data.status}`)}
                      </Badge>
                    </dd>
                  </div>
                  <div className="flex items-center justify-between gap-3">
                    <dt className="text-muted-foreground">
                      {t("shop.form.createdAtLabel")}
                    </dt>
                    <dd>{formatDateTime(profile.data.created_at)}</dd>
                  </div>
                </dl>
              ) : (
                <div className="flex flex-col gap-3">
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-2/3" />
                  <Skeleton className="h-4 w-1/2" />
                </div>
              )}
            </CardContent>
          </Card>
        ) : null}
      </div>
    </div>
  )
}
