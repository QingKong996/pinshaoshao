# 模拟支付平台

`internal/mockpayment` 是独立于 `internal/feature` 的开发用支付平台。它不引用任何业务 feature，业务侧通过 HTTP 接入；数据库只新增 `mock_payment_orders` 和 `mock_payment_notifications` 两张表。

前端位于 `frontend/src/mockpayment`，是 `frontend/src/features` 的同级独立模块，只复用 `frontend/src/shared`。

该模块用于本地开发和联调，不提供商户鉴权，也不处理真实资金。

## 配置

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `MOCKPAY_WEBHOOK_SECRET` | 是 | 回调 HMAC-SHA256 密钥 |
| `MOCKPAY_PUBLIC_BASE_URL` | 否 | 收银台公开地址前缀，例如 `http://localhost:8080`；为空时返回相对路径 |

启动前执行 `backend/migrations/20261124120000_create_mock_payments.sql` 迁移。

## 状态机

支付单只有以下两条合法状态边：

```text
pending ──pay───> paid
   │
   └──close───> closed
```

`paid` 和 `closed` 都是终态。重复支付已支付订单是幂等成功；已关闭订单不能支付；已支付订单不能关闭。超时任务和详情查询都会把到期的 `pending` 支付单关闭。

回调通知状态：

```text
pending ──claim──> delivering ──2xx──> succeeded
                         │
                         ├──可重试失败──> pending
                         └──达到上限────> exhausted
```

## HTTP API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/v1/mock-payments` | 创建支付单，`merchant_order_no` 幂等 |
| `GET` | `/api/v1/mock-payments/:paymentNo` | 查询支付单；到期时顺带关闭 |
| `POST` | `/api/v1/mock-payments/:paymentNo/close` | 主动关闭支付单 |
| `GET` | `/api/v1/mock-payments/:paymentNo/notifications` | 查询回调投递状态 |
| `POST` | `/api/v1/mock-payments/:paymentNo/notifications/retry` | 重置已耗尽的通知并立即重试 |
| `GET` | `/mock-payment/checkout/:paymentNo` | 收银台页面 |
| `POST` | `/mock-payment/checkout/:paymentNo/pay` | 收银台模拟支付 |
| `POST` | `/mock-payment/checkout/:paymentNo/close` | 收银台关闭支付单 |

创建支付单：

```bash
curl -X POST http://localhost:8080/api/v1/mock-payments \
  -H 'Content-Type: application/json' \
  -d '{
    "merchant_order_no": "order-20261124-001",
    "amount_cents": 3998,
    "currency": "CNY",
    "subject": "T-shirt order",
    "notify_url": "http://localhost:8080/api/v1/payment/webhook",
    "return_url": "http://localhost:5173/orders/1",
    "expires_in_seconds": 1800
  }'
```

响应中的 `checkout_url` 可直接交给浏览器。默认支付有效期 30 分钟，可通过 `expires_in_seconds` 设置为 60 秒至 24 小时。

## 前端控制台

| 路径 | 说明 |
| --- | --- |
| `/mock-payments` | 创建支付单、按支付单号查询 |
| `/mock-payments/:paymentNo` | 支付详情、显式状态机、收银台入口、主动关单、回调投递与人工重试 |

`/api` 与 `/mock-payment/checkout` 均由 Vite 代理到 `http://localhost:8080`。开发时先启动后端并执行迁移，再运行：

```bash
cd frontend
npm run dev
```

支付详情页在 `pending` 状态每 3 秒刷新，回调处于 `pending` 或 `delivering` 时每 2 秒刷新。前端不提供绕过收银台的支付接口，必须打开后端生成的 `checkout_url` 完成支付。

架构检查禁止 `src/features` 导入 `src/mockpayment`、禁止 `src/shared` 导入 `src/mockpayment`，并要求应用层只从 `src/mockpayment/index.ts` 公共入口导入。

## 回调协议

支付成功后，平台在同一数据库事务内写入回调 outbox，再由后台任务投递。每次回调包含：

| Header | 说明 |
| --- | --- |
| `X-MockPay-Event-Id` | 事件唯一 ID，重试期间不变 |
| `X-MockPay-Event-Type` | 当前为 `payment.paid` |
| `X-MockPay-Timestamp` | Unix 秒级时间戳 |
| `X-MockPay-Signature` | `v1=<hex>` |
| `X-MockPay-Delivery-Attempt` | 当前投递次数，从 1 开始 |

签名原文为：

```text
timestamp + "." + raw_request_body
```

签名值：

```text
v1=hex(HMAC-SHA256(MOCKPAY_WEBHOOK_SECRET, timestamp + "." + raw_request_body))
```

商户必须使用原始请求体验签，并校验时间戳窗口；示例实现可直接调用 `internal/mockpayment/security.Verify`。商户返回任意 `2xx` 视为成功，其他状态码或网络错误会进入重试。商户应按 `X-MockPay-Event-Id` 幂等处理回调。

默认最多投递 7 次，重试间隔依次为 1 秒、5 秒、30 秒、2 分钟、10 分钟、30 分钟。达到上限后通知进入 `exhausted`，可通过人工重试接口重新投递。

事件请求体示例：

```json
{
  "id": "a1c6cfe8-e77b-4ac2-b8bf-c6a0f56673f1",
  "type": "payment.paid",
  "occurred_at": "2026-11-24T12:00:00Z",
  "data": {
    "payment_no": "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
    "merchant_order_no": "order-20261124-001",
    "status": "paid",
    "amount_cents": 3998,
    "currency": "CNY",
    "subject": "T-shirt order",
    "paid_at": "2026-11-24T12:00:00Z"
  }
}
```

## 数据表

- `mock_payment_orders`：商户订单号、支付单号、金额、回调地址、有效期、支付/关闭时间和显式状态。
- `mock_payment_notifications`：回调事件 outbox、目标地址、原始 payload、投递次数、锁定租约、下次重试时间和错误。

两张表与业务订单表没有外键关系，因此模拟支付平台可单独部署或替换。

## 测试

```bash
go test ./internal/mockpayment/... ./internal/router
go test -tags integration ./internal/mockpayment -run TestMockPayment

cd frontend
npm run lint
npm run build
npm run test -- src/mockpayment
```
