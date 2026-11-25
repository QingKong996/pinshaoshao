# 模拟物流平台

`internal/mock/logistics` 是独立于 `internal/feature` 的开发用物流平台。它不引用任何业务 feature。它与主 API 运行在同一进程、同一端口，通过 `/mocklogistics` 路由命名空间隔离，并使用同一个 PostgreSQL 实例中的独立 `pinshaoshao_mocklogistics` 逻辑数据库。

模拟支付位于 `internal/mock/payment`，两者统一放在 `backend/internal/mock` 下。前端分别位于 `frontend/src/mock/payment` 和 `frontend/src/mock/logistics`。

该模块用于本地开发和联调，也不对接真实承运商。运营 API 使用独立 Bearer Session 鉴权。商家出库由控制台或 HTTP API 显式触发，之后按配置的总配送时长自动推进；业务物流接入见 [logistics.md](logistics.md)。

## 配置

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `MOCKLOGISTICS_DATABASE_URL` | 是 | `pinshaoshao_mocklogistics` 逻辑数据库 DSN，不回退到 `DATABASE_URL` |
| `MOCKLOGISTICS_WEBHOOK_SECRET` | 是 | 物流回调 HMAC-SHA256 密钥；为空时回退到 `MOCKPAY_WEBHOOK_SECRET` 以便本地联调 |
| `MOCKLOGISTICS_NOTIFY_ALLOWED_URLS` | 是 | 逗号分隔的回调 URL 白名单，必须包含完整路径，例如 `http://localhost:8080/api/v1/logistics/webhook` |
| `MOCKLOGISTICS_DELIVERY_DURATION` | 否 | 商家出库到签收的总时长，Go duration 格式；默认 `4m`，最大 `5m`，超过上限会截断到 `5m` |
| `MOCKLOGISTICS_AUTH_USERNAME` | 是 | MockLogistics 运营账号 |
| `MOCKLOGISTICS_AUTH_PASSWORD` | 是 | MockLogistics 运营密码；仅以 Argon2id 哈希落库 |
| `MOCKLOGISTICS_AUTH_SESSION_TTL` | 否 | Session 有效期，默认 `24h` |

无法解析或非正数的配送时长会回退到默认值。

`notify_url` 必须与 `MOCKLOGISTICS_NOTIFY_ALLOWED_URLS` 中的一项完全匹配，包括 scheme、host、port 和 path；不接受 query、fragment、userinfo、点段或编码路径分隔符。创建物流单和后台实际投递时都会校验，白名单外或配置无效时启动失败。

数据库创建、迁移和隔离校验见 [mock-databases.md](mock-databases.md)。

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

- 商家出库后，后台按总配送时长平均推进 5 个节点；默认 4 分钟完成，进程重启后仍按 `dispatched_at` 继续推进。
- 总配送时长由 `MOCKLOGISTICS_DELIVERY_DURATION` 配置，例如 `2m`、`90s`；大于 `5m` 的值会截断到 `5m`。
- 创建后为 `pending_outbound`，不能直接推进，必须先由商家确认出库。
- `outbound` 是商家动作；重复调用已出库的物流单会返回当前状态，不重复写事件或回调。
- `advance` 每次只推进一个节点，可用来加速自动流程；`out_for_delivery` 不能继续 `advance`。
- `complete` 用于在派送中提前签收；自动流程到达总配送时长后会自行签收。重复完成已签收物流单是幂等成功。
- 创建只写初始轨迹，不发送回调；从商家出库开始，自动或手动状态变化都会写一条物流事件，并在同一事务内写入一条回调通知。

## Mock 数据

使用 `github.com/brianvoe/gofakeit/v7` 生成：

- 承运商、运单号前缀和商品摘要；
- 寄件人、收件人、电话和地址；
- 揽收快递员、派送快递员；
- 未指定时的始发城市和目的城市。

商家订单号、回调地址和调用方显式提供的字段不会被随机覆盖。相同商户订单号和请求参数重放时返回原物流单。

内置本地回调接收器：`POST http://localhost:8080/mocklogistics/webhook`，返回 `204`，便于在控制台直接看到回调投递成功。接入真实服务时替换 `notify_url`，并把完整回调 URL 加入 `MOCKLOGISTICS_NOTIFY_ALLOWED_URLS`。

## HTTP API

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/mocklogistics/api/v1/auth/login` | 公开 | 用户名和密码换取 Session token |
| `POST` | `/mocklogistics/api/v1/auth/logout` | Bearer Session | 撤销当前 Session |
| `POST` | `/mocklogistics/api/v1/shipments` | Bearer Session | 创建物流单，`merchant_order_no` 幂等 |
| `GET` | `/mocklogistics/api/v1/shipments/:shipmentNo` | Bearer Session | 查询物流单、当前状态和完整轨迹 |
| `POST` | `/mocklogistics/api/v1/shipments/:shipmentNo/outbound` | Bearer Session | 商家确认出库 |
| `POST` | `/mocklogistics/api/v1/shipments/:shipmentNo/advance` | Bearer Session | 手动推进到下一物流节点，可加速自动流程 |
| `POST` | `/mocklogistics/api/v1/shipments/:shipmentNo/complete` | Bearer Session | 手动提前签收；自动流程会按时签收 |
| `GET` | `/mocklogistics/api/v1/shipments/:shipmentNo/notifications` | Bearer Session | 查询回调投递记录 |
| `POST` | `/mocklogistics/api/v1/shipments/:shipmentNo/notifications/:notificationID/retry` | Bearer Session | 重置已耗尽的通知并立即重试 |
| `POST` | `/mocklogistics/webhook` | 公开空接收端 | 内置回调接收器 |

创建物流单：

```bash
curl -X POST http://localhost:8080/mocklogistics/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"<MOCKLOGISTICS_AUTH_USERNAME>","password":"<MOCKLOGISTICS_AUTH_PASSWORD>"}'

export TOKEN='<token returned by login>'
curl -X POST http://localhost:8080/mocklogistics/api/v1/shipments \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "merchant_order_no": "order-20261124-001",
    "item_summary": "Wireless keyboard",
    "carrier": "顺丰速运",
    "origin_city": "上海",
    "destination_city": "杭州",
    "notify_url": "http://localhost:8080/mocklogistics/webhook"
  }'
```

手动推进完整流程（可选，自动流程无需调用）：

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/outbound
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/advance
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/advance
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/advance
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/advance
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/complete
```

## 前端控制台

| 路径 | 说明 |
| --- | --- |
| `/mocklogistics/signin` | MockLogistics 运营登录 |
| `/mocklogistics` | 创建物流单、按物流单号查询；需要运营 Session |
| `/mocklogistics/shipments/:shipmentNo` | 状态、商家出库、自动/手动推进、完成、轨迹和回调记录；需要运营 Session |

`/mocklogistics/api` 由 Vite 代理到 `http://localhost:8080`。前端路由本身不代理。

MockLogistics 前端 Session 使用独立于 MockPay、买家 Session 和商户 Session 的 `sessionStorage` 键。业务物流 Gateway 复用同一运营账号登录，并在收到 `401` 后刷新 Session、重试一次请求。

## 回调协议

物流状态变化后，平台在同一数据库事务内写入回调 outbox，再由后台任务投递。每次回调包含：

| Header | 说明 |
| --- | --- |
| `X-MockLogistics-Event-Id` | 事件唯一 ID，重试期间不变 |
| `X-MockLogistics-Event-Type` | 例如 `shipment.dispatched`、`shipment.delivered` |
| `X-MockLogistics-Timestamp` | Unix 秒级时间戳 |
| `X-MockLogistics-Signature` | `v1=<hex>` |
| `X-MockLogistics-Delivery-Attempt` | 当前投递次数，从 1 开始 |

签名原文：

```text
timestamp + "." + raw_request_body
```

签名值：

```text
v1=hex(HMAC-SHA256(MOCKLOGISTICS_WEBHOOK_SECRET, timestamp + "." + raw_request_body))
```

验签算法复用中立的 `internal/webhooksig`。商户返回任意 `2xx` 视为成功，其他状态码或网络错误进入重试。投递失败时 `last_error` 只记录状态码或固定摘要，不保存目标响应体或请求 URL。默认最多投递 7 次，退避间隔为 1 秒、5 秒、30 秒、2 分钟、10 分钟、30 分钟；耗尽后可人工重试。

事件示例：

```json
{
  "id": "a1c6cfe8-e77b-4ac2-b8bf-c6a0f56673f1",
  "type": "shipment.delivered",
  "occurred_at": "2026-11-24T12:00:00Z",
  "data": {
    "shipment_no": "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
    "tracking_no": "SF123456789012",
    "merchant_order_no": "order-20261124-001",
    "status": "delivered",
    "carrier": "顺丰速运",
    "sequence": 7,
    "description": "快件已签收，感谢使用",
    "location": "杭州",
    "occurred_at": "2026-11-24T12:00:00Z"
  }
}
```

## 数据表

- `mock_shipments`：物流单、商家订单号、运单号、承运商、收寄件人、城市、状态和时间戳。
- `mock_logistics_events`：物流轨迹，按物流单和 sequence 唯一。
- `mock_logistics_notifications`：回调 outbox、目标地址、原始 payload、投递次数、锁定租约、下次重试时间和错误。
- `mock_auth_accounts`：环境变量引导的运营账号及 Argon2id 密码哈希。
- `mock_auth_sessions`：32 字节 Session token 的 SHA-256 哈希、账号关联、创建时间和过期时间。

这些表与业务表没有外键关系，也不存在跨数据库查询。MockLogistics 使用独立的 `pinshaoshao_mocklogistics` 逻辑数据库。

## 边界

- 禁止 `internal/feature` 导入 `internal/mock`，也禁止模拟模块导入业务 feature。
- 模拟支付与模拟物流也不互相导入，二者只共享 `internal/webhooksig` 这类中立基础设施。
- 前端应用层只从 `src/mock/logistics/index.ts` 公共入口导入，业务 feature 不依赖模拟模块。
- 模拟物流本身不改商城订单状态。业务 `logistics` feature 通过 HTTP 创建并出库物流单，再通过签名回调推进子订单与父订单；模拟控制台仍可直接操作物流平台用于联调。
- 自动推进 worker 与通知 worker 随模块 `Start` 启动，只查询 `mock_shipments` 并调用同一状态机，不绕过事件和回调事务。

## 测试

```bash
go test ./internal/mock/... ./internal/http/sessionclient ./internal/router ./internal/architecture
go test -tags integration ./internal/mock/logistics -run TestMockLogistics

cd frontend
npm run lint
npm run build
```
