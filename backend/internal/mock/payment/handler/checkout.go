package handler

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"pinshaoshao/internal/mock/payment/model"

	"github.com/gin-gonic/gin"
)

var checkoutTemplate = template.Must(template.New("checkout").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Mock Payment</title>
  <style>
    :root { color-scheme: light; --ink: #17211b; --muted: #617068; --accent: #146c43; --line: #dce4df; --paper: #f5f7f4; }
    * { box-sizing: border-box; }
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; padding: 24px; background: var(--paper); color: var(--ink); font-family: ui-sans-serif, system-ui, sans-serif; }
    main { width: min(100%, 520px); background: white; border: 1px solid var(--line); border-radius: 20px; padding: 28px; box-shadow: 0 24px 70px rgb(23 33 27 / 10%); }
    .eyebrow { color: var(--accent); font-size: 12px; font-weight: 800; letter-spacing: .12em; text-transform: uppercase; }
    h1 { margin: 8px 0 4px; font-size: 28px; }
    .merchant { color: var(--muted); margin: 0 0 24px; }
    dl { display: grid; grid-template-columns: 120px 1fr; gap: 12px; margin: 0 0 24px; }
    dt { color: var(--muted); }
    dd { margin: 0; font-weight: 650; overflow-wrap: anywhere; }
    .amount { font-size: 32px; letter-spacing: -.03em; }
    .status { display: inline-flex; padding: 5px 10px; border-radius: 999px; background: #e8f2ec; color: var(--accent); font-size: 13px; font-weight: 750; }
    .status.closed { background: #f1f1ef; color: #686863; }
    .error { padding: 12px 14px; margin-bottom: 20px; border-radius: 10px; background: #fff1ed; color: #9b2c14; }
    .actions { display: flex; gap: 10px; flex-wrap: wrap; }
    button, a.button { appearance: none; border: 0; border-radius: 10px; padding: 12px 18px; font: inherit; font-weight: 750; text-decoration: none; cursor: pointer; }
    button.primary, a.button.primary { background: var(--accent); color: white; }
    button.secondary { background: #edf1ee; color: var(--ink); }
    .note { color: var(--muted); font-size: 13px; line-height: 1.5; margin: 20px 0 0; }
  </style>
</head>
<body>
<main>
  <div class="eyebrow">Mock payment platform</div>
  <h1>{{.Subject}}</h1>
  <p class="merchant">Merchant order {{.MerchantOrderNo}}</p>
  {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
  <dl>
    <dt>Amount</dt><dd class="amount">{{.Amount}}</dd>
    <dt>Status</dt><dd><span class="status {{if eq .Status "closed"}}closed{{end}}">{{.Status}}</span></dd>
    <dt>Payment no.</dt><dd>{{.PaymentNo}}</dd>
    <dt>Expires</dt><dd>{{.ExpiresAt}}</dd>
  </dl>
  {{if .CanPay}}
  <div class="actions">
    <form method="post" action="/mockpay/checkout/{{.PaymentNo}}/pay"><button class="primary" type="submit">Pay now</button></form>
    <form method="post" action="/mockpay/checkout/{{.PaymentNo}}/close"><button class="secondary" type="submit">Close order</button></form>
  </div>
  {{else if eq .Status "paid"}}
  <div class="actions">
    {{if .ReturnURL}}<a class="button primary" href="{{.ReturnURL}}">Return to merchant</a>{{end}}
  </div>
  <p class="note">The signed payment notification is queued for delivery. Repeated callbacks use the same event id.</p>
  {{else}}
  <p class="note">This payment order is closed and can no longer be paid.</p>
  {{end}}
</main>
</body>
</html>`))

var checkoutErrorTemplate = template.Must(template.New("checkout-error").Parse(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Mock Payment</title></head>
<body style="font-family: ui-sans-serif, system-ui, sans-serif; max-width: 560px; margin: 80px auto; padding: 0 24px;">
  <h1>Mock payment</h1>
  <p>{{.}}</p>
</body>
</html>`))

type CheckoutService interface {
	Get(ctx context.Context, paymentNo string) (*model.Payment, error)
	Pay(ctx context.Context, paymentNo string) (*model.Payment, error)
	Close(ctx context.Context, paymentNo string) (*model.Payment, error)
}

type CheckoutHandler struct {
	service CheckoutService
}

func NewCheckoutHandler(checkoutService CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{service: checkoutService}
}

type checkoutView struct {
	PaymentNo       string
	MerchantOrderNo string
	Subject         string
	Amount          string
	Status          string
	ExpiresAt       string
	ReturnURL       string
	CanPay          bool
	Error           string
}

func (h *CheckoutHandler) Show(ctx *gin.Context) {
	payment, err := h.service.Get(ctx.Request.Context(), ctx.Param("paymentNo"))
	if err != nil {
		h.renderError(ctx, err)
		return
	}
	h.render(ctx, http.StatusOK, payment, "")
}

func (h *CheckoutHandler) Pay(ctx *gin.Context) {
	payment, err := h.service.Pay(ctx.Request.Context(), ctx.Param("paymentNo"))
	if err != nil {
		h.renderActionError(ctx, err)
		return
	}
	ctx.Redirect(http.StatusSeeOther, checkoutPathPrefix+payment.PaymentNo)
}

func (h *CheckoutHandler) Close(ctx *gin.Context) {
	payment, err := h.service.Close(ctx.Request.Context(), ctx.Param("paymentNo"))
	if err != nil {
		h.renderActionError(ctx, err)
		return
	}
	ctx.Redirect(http.StatusSeeOther, checkoutPathPrefix+payment.PaymentNo)
}

func (h *CheckoutHandler) renderActionError(ctx *gin.Context, actionErr error) {
	payment, err := h.service.Get(ctx.Request.Context(), ctx.Param("paymentNo"))
	if err != nil {
		h.renderError(ctx, err)
		return
	}

	status, message := httpStatusForError(actionErr)
	if status == http.StatusInternalServerError {
		slog.ErrorContext(ctx.Request.Context(), "mock payment checkout action", "error", actionErr)
		message = "The payment action could not be completed."
	}
	h.render(ctx, status, payment, message)
}

func (h *CheckoutHandler) render(ctx *gin.Context, status int, payment *model.Payment, message string) {
	view := checkoutView{
		PaymentNo:       payment.PaymentNo,
		MerchantOrderNo: payment.MerchantOrderNo,
		Subject:         payment.Subject,
		Amount:          formatAmount(payment.AmountCents, payment.Currency),
		Status:          string(payment.Status),
		ExpiresAt:       payment.ExpiresAt.UTC().Format("2006-01-02 15:04:05 MST"),
		ReturnURL:       payment.ReturnURL,
		CanPay:          payment.Status == model.PaymentStatusPending,
		Error:           message,
	}
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Status(status)
	if err := checkoutTemplate.Execute(ctx.Writer, view); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "render mock payment checkout", "error", err)
	}
}

func (h *CheckoutHandler) renderError(ctx *gin.Context, err error) {
	status, message := httpStatusForError(err)
	if status == http.StatusInternalServerError {
		slog.ErrorContext(ctx.Request.Context(), "load mock payment checkout", "error", err)
		message = "The payment order could not be loaded."
	}
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Status(status)
	if renderErr := checkoutErrorTemplate.Execute(ctx.Writer, message); renderErr != nil {
		slog.ErrorContext(ctx.Request.Context(), "render mock payment checkout error", "error", renderErr)
	}
}

func formatAmount(amountCents int64, currency string) string {
	return fmt.Sprintf("%s %d.%02d", currency, amountCents/100, amountCents%100)
}
