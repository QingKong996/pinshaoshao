# 模拟物流平台

`internal/mock/logistics` 是独立于 `internal/feature` 的开发用物流平台。它不引用任何业务 feature，数据库只新增 `mock_shipments`、`mock_logistics_events` 和 `mock_logistics_notifications`。它与主 API 运行在同一进程、同一端口，通过 `/mocklogistics` 路由命名空间隔离。

模拟支付位于 `internal/mock/payment`，两者统一放在 `backend/internal/mock` 下。前端分别位于 `frontend/src/mock/payment` 和 `frontend/src/mock/logistics`。

该模块用于本地开发和联调，不提供鉴权，也不对接真实承运商。商家出库由控制台或 HTTP API 显式触发；业务物流接入见 [logistics.md](logistics.md)。

## 配置

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `MOCKLOGISTICS_WEBHOOK_SECRET` | 是 | 物流回调 HMAC-SHA256 密钥；为空时回退到 `MOCKPAY_WEBHOOK_SECRET` 以便本地联调 |

启动前执行 `backend/migrations/20261124140000_create_mock_logistics.sql` 迁移。

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
- `outbound` 是商家动作；重复调用已出库的物流单会返回当前状态，不重复写事件或回调。
- `advance` 每次只推进一个节点；`out_for_delivery` 不能继续 `advance`，必须调用 `complete`。
- `complete` 只在 `out_for_delivery` 可用；重复完成已签收物流单是幂等成功。
- 创建只写初始轨迹，不发送回调；从商家出库开始，每次真实状态变化都会写一条物流事件，并在同一事务内写入一条回调通知。

## Mock 数据

使用 `github.com/brianvoe/gofakeit/v7` 生成：

- 承运商、运单号前缀和商品摘要；
- 寄件人、收件人、电话和地址；
- 揽收快递员、派送快递员；
- 未指定时的始发城市和目的城市。

商家订单号、回调地址和调用方显式提供的字段不会被随机覆盖。相同商户订单号和请求参数重放时返回原物流单。

内置本地回调接收器：`POST http://localhost:8080/mocklogistics/webhook`，返回 `204`，便于在控制台直接看到回调投递成功。接入真实服务时替换 `notify_url`。

## HTTP API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/mocklogistics/api/v1/shipments` | 创建物流单，`merchant_order_no` 幂等 |
| `GET` | `/mocklogistics/api/v1/shipments/:shipmentNo` | 查询物流单、当前状态和完整轨迹 |
| `POST` | `/mocklogistics/api/v1/shipments/:shipmentNo/outbound` | 商家确认出库 |
| `POST` | `/mocklogistics/api/v1/shipments/:shipmentNo/advance` | 推进到下一物流节点 |
| `POST` | `/mocklogistics/api/v1/shipments/:shipmentNo/complete` | 确认签收并完成 |
| `GET` | `/mocklogistics/api/v1/shipments/:shipmentNo/notifications` | 查询回调投递记录 |
| `POST` | `/mocklogistics/api/v1/shipments/:shipmentNo/notifications/:notificationID/retry` | 重置已耗尽的通知并立即重试 |
| `POST` | `/mocklogistics/webhook` | 内置回调接收器 |

创建物流单：

```bash
curl -X POST http://localhost:8080/mocklogistics/api/v1/shipments \
  -H 'Content-Type: application/json' \
  -d '{
    "merchant_order_no": "order-20261124-001",
    "item_summary": "Wireless keyboard",
    "carrier": "顺丰速运",
    "origin_city": "上海",
    "destination_city": "杭州",
    "notify_url": "http://localhost:8080/mocklogistics/webhook"
  }'
```

推进完整流程：

```bash
curl -X POST http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/outbound
curl -X POST http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/advance
curl -X POST http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/advance
curl -X POST http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/advance
curl -X POST http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/advance
curl -X POST http://localhost:8080/mocklogistics/api/v1/shipments/{shipmentNo}/complete
```

## 前端控制台

| 路径 | 说明 |
| --- | --- |
| `/mocklogistics` | 创建物流单、按物流单号查询 |
| `/mocklogistics/shipments/:shipmentNo` | 状态、商家出库、推进、完成、轨迹和回调记录 |

`/mocklogistics/api` 由 Vite 代理到 `http://localhost:8080`。前端路由本身不代理。

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

验签算法复用中立的 `internal/webhooksig`。商户返回任意 `2xx` 视为成功，其他状态码或网络错误进入重试。默认最多投递 7 次，退避间隔为 1 秒、5 秒、30 秒、2 分钟、10 分钟、30 分钟；耗尽后可人工重试。

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

三张表与业务订单表没有外键关系。

## 边界

- 禁止 `internal/feature` 导入 `internal/mock`，也禁止模拟模块导入业务 feature。
- 模拟支付与模拟物流也不互相导入，二者只共享 `internal/webhooksig` 这类中立基础设施。
- 前端应用层只从 `src/mock/logistics/index.ts` 公共入口导入，业务 feature 不依赖模拟模块。
- 模拟物流本身不改商城订单状态。业务 `logistics` feature 通过 HTTP 创建并出库物流单，再通过签名回调推进子订单与父订单；模拟控制台仍可直接操作物流平台用于联调。

## 测试

```bash
go test ./internal/mock/... ./internal/router ./internal/architecture
go test -tags integration ./internal/mock/logistics -run TestMockLogisticsFlow

cd frontend
npm run lint
npm run build
```
