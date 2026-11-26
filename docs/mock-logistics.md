# 模拟物流平台

`internal/mock/logistics` 是开发用物流平台，不引用任何业务 feature，也不对接真实承运商。它与主 API 同进程、同端口，通过 `/mocklogistics` 路由隔离，并使用同一 PostgreSQL 实例中的独立 `pinshaoshao_mocklogistics` 逻辑数据库。

业务物流接入见 [logistics.md](logistics.md)。

## 访问边界

- 运营 API 使用独立 Bearer Session 鉴权。
- 业务物流 Gateway 通过中立的 `sessionclient` 登录并缓存 Session，收到 `401` 后刷新并重试一次。
- 商家出库由控制台或 HTTP API 显式触发；之后按配送总时长自动推进。

## 状态机

```text
pending_outbound
  │ merchant outbound
  ▼
pending_pickup
  │ advance
  ▼
picked_up
  │ advance
  ▼
in_transit
  │ advance
  ▼
at_transit_hub
  │ advance
  ▼
out_for_delivery
  │ complete
  ▼
delivered
```

- 创建后为 `pending_outbound`，不能直接推进，必须先由商家确认出库。
- 出库后，后台按总配送时长平均推进 5 个节点；默认 4 分钟完成，进程重启后仍按 `dispatched_at` 继续推进。
- `advance` 每次只推进一个节点，可用来加速自动流程；`out_for_delivery` 不能继续 `advance`。
- `complete` 用于提前签收；自动流程到达总配送时长后也会自行签收。
- 重复出库或重复完成已签收物流单是幂等成功，不重复写事件或回调。

## 轨迹与回调

- 创建物流单只写初始轨迹，不发送回调。
- 从商家出库开始，每次状态变化都会在同一事务内写入物流事件和回调 outbox。
- 回调头为 `X-MockLogistics-Event-Id`、`X-MockLogistics-Event-Type`、`X-MockLogistics-Timestamp`、`X-MockLogistics-Signature` 和 `X-MockLogistics-Delivery-Attempt`。
- 签名原文为 `timestamp + "." + raw_request_body`，签名算法位于中立的 `internal/webhooksig`。
- 业务侧返回任意 `2xx` 视为成功；其他状态码或网络错误进入重试。事件 ID 在重试期间保持不变。
- 回调中的 `merchant_order_no` 是业务发货单号，`shipment_no` 是物流平台发货单号，`sequence` 用于业务侧检测轨迹冲突。

回调 outbox 的租约领取、退避重试、耗尽后人工重试和错误字段脱敏规则与 [mock-payment.md](mock-payment.md) 一致。

## Mock 数据

使用 `github.com/brianvoe/gofakeit/v7` 生成承运商、运单号前缀、商品摘要、寄件人、收件人、电话、地址和快递员信息。商家订单号、回调地址和调用方显式提供的字段不会被随机覆盖；相同商户订单号和请求参数重放时返回原物流单。

内置本地回调接收器 `POST /mocklogistics/webhook` 返回 `204`，便于在控制台直接观察投递成功。

## 数据与边界

- `mock_shipments`：物流单、商家订单号、运单号、承运商、收寄件人、城市、状态和时间戳。
- `mock_logistics_events`：物流轨迹，按物流单和 sequence 唯一。
- `mock_logistics_notifications`：回调 outbox、目标地址、原始 payload、投递次数、锁定租约、下次重试时间和错误。
- `mock_auth_accounts` 与 `mock_auth_sessions`：环境变量引导的运营账号，以及只保存哈希的 Session。
- 这些表属于 `pinshaoshao_mocklogistics`，与业务表没有外键关系，也不存在跨数据库查询。
- 模拟物流不改商城订单状态；业务 `logistics` feature 通过 HTTP 创建并出库物流单，再通过签名回调推进订单。
- 自动推进 worker 与通知 worker 随模块启动，只查询模拟物流表并调用同一状态机，不绕过事件和回调事务。
- 模拟物流与模拟支付、业务 feature 都不互相 import。

数据库创建与迁移见 [mock-databases.md](mock-databases.md)，配置见 `backend/.env.example`。
