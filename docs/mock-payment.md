# 模拟支付平台

`internal/mock/payment` 是开发用支付平台，不引用任何业务 feature，也不处理真实资金。它与主 API 同进程、同端口，通过 `/mockpay` 路由隔离，并使用同一 PostgreSQL 实例中的独立 `pinshaoshao_mockpay` 逻辑数据库。

业务支付接入见 [payment.md](payment.md)。

## 访问边界

- 运营 API 使用独立 Bearer Session 鉴权。
- 收银台使用高熵 `paymentNo` 作为能力凭证，不要求运营登录。
- 业务支付 Gateway 通过中立的 `sessionclient` 登录并缓存 Session，收到 `401` 后刷新并重试一次。
- 收银台由后端提供，模拟控制台不能绕过收银台直接标记支付成功。

## 状态机

```text
pending ──pay───> paid
   │
   └──close───> closed
```

`paid` 和 `closed` 都是终态。重复支付已支付订单是幂等成功；已关闭订单不能支付；已支付订单不能关闭。过期任务和详情查询都会把到期的 `pending` 支付单关闭。

回调通知状态：

```text
pending ──claim──> delivering ──2xx──> succeeded
                         │
                         ├──可重试失败──> pending
                         └──达到上限────> exhausted
```

## 回调投递

支付成功后，平台在同一数据库事务内写入回调 outbox，再由后台 worker 投递。事件 ID 在重试期间保持不变，可按它幂等处理。

每次回调包含：

- `X-MockPay-Event-Id`
- `X-MockPay-Event-Type: payment.paid`
- `X-MockPay-Timestamp`
- `X-MockPay-Signature: v1=<hex>`
- `X-MockPay-Delivery-Attempt`

签名原文为 `timestamp + "." + raw_request_body`：

```text
v1=hex(HMAC-SHA256(MOCKPAY_WEBHOOK_SECRET, timestamp + "." + raw_request_body))
```

商户返回任意 `2xx` 视为成功；其他状态码或网络错误进入重试。投递失败时 `last_error` 只记录状态码或固定摘要，不保存目标响应体或请求 URL。

默认最多投递 7 次，退避间隔为 1 秒、5 秒、30 秒、2 分钟、10 分钟、30 分钟。达到上限后通知进入 `exhausted`，可通过运营接口重置并立即重试。

`notify_url` 必须与 `MOCKPAY_NOTIFY_ALLOWED_URLS` 中的一项完全匹配，包括 scheme、host、port 和 path；不接受 query、fragment、userinfo、点段或编码路径分隔符。创建支付单和实际投递时都会校验，白名单外或配置无效时启动失败。

## 数据与边界

- `mock_payment_orders`：商户订单号、支付单号、金额、回调地址、有效期、支付/关闭时间和显式状态。
- `mock_payment_notifications`：回调 outbox、目标地址、原始 payload、投递次数、锁定租约、下次重试时间和错误。
- `mock_auth_accounts` 与 `mock_auth_sessions`：环境变量引导的运营账号，以及只保存哈希的 Session。
- 这些表属于 `pinshaoshao_mockpay`，与业务表没有外键关系，也不存在跨数据库查询。
- 模拟支付与业务支付不互相 import；共享的验签算法位于中立的 `internal/webhooksig`。
- 模拟支付与模拟物流也不互相 import。

数据库创建与迁移见 [mock-databases.md](mock-databases.md)，配置见 `backend/.env.example`。
