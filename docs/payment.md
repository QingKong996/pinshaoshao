# 业务支付

`internal/feature/payment` 是商城侧的支付 feature。它不实现收银台，也不保存渠道内部通知队列；渠道能力由同一 API 进程中的 `internal/mock/payment` 提供，双方仍只通过网络通信。

支付 feature 与模拟支付平台不互相 import，只能通过 HTTP API 和签名回调通信。业务 Gateway 复用中立的 `internal/http/sessionclient` 管理 MockPay Bearer Session，双方共享的算法位于中立 `internal/webhooksig`。

## 职责边界

- `payment` 拥有 `payments` 表，每张父订单最多一张业务支付单。
- `payment_no` 是业务支付单号，同时作为向渠道创建支付会话的幂等键。
- 创建业务支付单只落本地记录；发起支付时才向渠道创建或复用支付会话。
- 渠道支付成功后发送签名回调；`payment` 验签、核对金额与流水，并在同一事务内推进支付单、父订单和子订单状态。
- 支付模块只提供关单操作，不允许修改金额或流水。状态只能按 `pending -> succeeded` 或 `pending -> closed` 流转。

## 环境变量

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `MOCKPAY_WEBHOOK_SECRET` | 是 | 与 mockpayment 共用的回调 HMAC 密钥 |
| `PAYMENT_PUBLIC_BASE_URL` | 否 | 本服务公开地址，默认 `http://localhost:$PORT` |
| `PAYMENT_GATEWAY_BASE_URL` | 否 | MockPay 服务 origin，不包含 `/mockpay/api/v1`；默认同 `PAYMENT_PUBLIC_BASE_URL` |
| `PAYMENT_NOTIFY_URL` | 否 | 回调地址，默认 `$PAYMENT_PUBLIC_BASE_URL/api/v1/payment/webhook` |
| `PAYMENT_RETURN_URL_BASE` | 否 | 渠道支付完成后的前端地址，默认 `http://localhost:5173` |
| `MOCKPAY_NOTIFY_ALLOWED_URLS` | 是 | MockPay 允许投递的回调 URL；必须包含完整的 `PAYMENT_NOTIFY_URL` |
| `MOCKPAY_AUTH_USERNAME` | 是 | MockPay 运营账号；业务 Gateway 使用同一账号登录 |
| `MOCKPAY_AUTH_PASSWORD` | 是 | MockPay 运营密码 |

启动前执行 `backend/migrations/20261124130000_create_payments.sql`。

## HTTP API

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/api/v1/payments` | 买家会话 | 为订单创建或返回已有业务支付单 |
| `GET` | `/api/v1/payments/:paymentNo` | 买家会话 | 查询自己的支付单 |
| `POST` | `/api/v1/payments/:paymentNo/pay` | 买家会话 | 创建或复用渠道支付会话，返回 `checkout_url` |
| `POST` | `/api/v1/payments/:paymentNo/close` | 买家会话 | 关闭待支付支付单，同时取消订单并释放库存 |
| `POST` | `/api/v1/payment/webhook` | 签名 | 接收 `payment.paid` 回调 |

创建支付单：

```http
POST /api/v1/payments
Authorization: Bearer <token>
Content-Type: application/json

{"order_id": 1}
```

首次创建返回 `201`；相同订单重复调用返回原支付单和 `200`。

发起支付：

```http
POST /api/v1/payments/4f2614df-f674-4b5e-a9e4-3fc5e7f62013/pay
Authorization: Bearer <token>
```

响应中的 `checkout_url` 交给浏览器完成支付。重复调用使用同一 `payment_no` 请求渠道，不会创建第二张渠道支付单。

关闭支付单：

```http
POST /api/v1/payments/4f2614df-f674-4b5e-a9e4-3fc5e7f62013/close
Authorization: Bearer <token>
```

仅支付单所有者可以关闭。已有渠道流水时先请求 MockPay 关单，成功后在本地事务中取消父订单、取消子订单、释放冻结库存，并将支付单置为 `closed`。重复关闭返回 `closed`；已支付支付单返回 `409`。

## 状态与一致性

业务支付单状态：

```text
pending ──callback──> succeeded
   │
   └──close───────> closed
```

- 创建支付单时锁定父订单。订单不属于当前买家、已支付、已取消或已过期时拒绝创建。
- 支付成功回调先锁父订单，再锁支付单。仅 `pending_payment` 订单推进为 `paid`，其子订单推进为 `awaiting_shipment`。
- 主动关单同样先锁父订单，再锁支付单；只有 `pending_payment` 订单会被取消并释放库存，重复关单保持幂等。
- 渠道已支付时关单返回 `409`，本地订单不变，支付回调仍可正常结算订单。
- 相同渠道流水的重复回调直接成功，不重复结算。
- 回调金额与业务支付单金额不一致、渠道流水与已记录流水不一致时整笔回滚，不改变订单状态。
- 订单已取消后收到支付成功回调时，只记录支付成功和渠道流水，不恢复订单、不重新冻结库存。
- 渠道支付成功后由 mockpayment 的 outbox worker 投递通知；失败重试与人工重试由渠道侧负责。

## 回调协议

回调头与签名原文和 mockpayment 文档一致：

- `X-MockPay-Event-Id`
- `X-MockPay-Event-Type: payment.paid`
- `X-MockPay-Timestamp`
- `X-MockPay-Signature: v1=<hex>`

签名原文为 `timestamp + "." + raw_request_body`。业务侧先使用原始请求体验签，再解析 JSON；默认允许 5 分钟时间偏差。

回调中的 `merchant_order_no` 对应业务支付单号，`payment_no` 对应渠道流水号。处理成功返回 `204`。

## 前端

买家支付入口位于订单详情页 `/orders/:id`，实现代码在 `frontend/src/features/payment`。

- 待支付订单在右侧显示支付卡片和“立即支付”按钮。
- 点击后依次调用创建支付单和发起支付接口，取得 `checkout_url` 后跳转 MockPay 收银台。
- 支付卡片提供“关闭支付并取消订单”；二次确认后先幂等获取支付单，再调用关单接口并刷新订单。
- 支付完成后回到订单详情页；待支付订单每 3 秒轮询一次，回调结算后自动显示“已支付”。
- 支付成功或订单已支付时显示支付完成状态；取消订单和后续履约状态不显示支付卡片。

`PAYMENT_RETURN_URL_BASE` 必须与买家访问前端时使用的 origin 完全一致。买家会话保存在 `localStorage`，`localhost` 与 `127.0.0.1` 属于不同 origin，混用会导致支付返回后丢失登录状态。

## 测试

```bash
cd backend
go test ./internal/feature/payment/... ./internal/router
go test -tags integration ./internal/feature/payment/repository
go test -tags integration ./internal/e2e

cd ../frontend
npm run test -- src/features/payment
npm run test:e2e -- e2e/buyer-payment.spec.ts
npm run lint
npm run build
```

`internal/e2e` 使用真实 PostgreSQL 与真实 HTTP 请求覆盖：直接下单、创建业务支付单、调用 MockPay API、收银台支付、签名回调、订单与子订单结算、通知投递完成。该测试中业务支付与 MockPay 仍只通过网络通信。
