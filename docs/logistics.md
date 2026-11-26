# 业务物流

`internal/feature/logistics` 拥有商城侧发货单与物流轨迹。承运、轨迹推进和回调重试由独立的 `internal/mock/logistics` 平台提供，双方只通过 HTTP API 与签名回调通信。模拟物流平台见 [mock-logistics.md](mock-logistics.md)。

## 边界与状态

- 业务侧拥有 `shipments` 与 `shipment_events`；每个子订单最多一张发货单。
- 商家重复发货返回已有发货单，不会创建第二张业务发货单。
- 发货单初始状态只允许为 `pending_pickup`。
- 业务 `logistics` 不直接访问订单表或库存表，只通过 `order/shippingport` 暴露的履约端口协作。
- `order` 不依赖 `logistics`；组合根把物流生命周期端口注入物流模块。

发货单状态：

```text
pending_pickup -> picked_up -> in_transit -> at_transit_hub -> out_for_delivery -> delivered
```

子订单与父订单状态：

```text
awaiting_shipment --ship--> shipped --deliver--> completed
```

- 部分子订单发货时，父订单为 `partially_shipped`。
- 所有子订单发货后，父订单为 `shipped`。
- 所有子订单签收后，父订单为 `completed`。

## 一致性与并发

- 商家发货先调用 MockLogistics 创建物流单并确认出库，再开启本地事务。
- 业务发货单号作为物流平台的 `merchant_order_no` 幂等键；本地事务失败后重试会复用平台已有物流单。
- 本地事务锁定订单、子订单与库存，在同一事务中扣减冻结库存、写入发货单、推进子订单并重新计算父订单。
- 发货事务与签收回调统一先锁订单，再锁发货单，避免并发时形成锁环。
- 物流平台在业务发货单落库前投递回调时可能收到 `404`；平台按自身退避策略重试，业务单落库后即可成功。

## 回调与幂等

回调头：

- `X-MockLogistics-Event-Id`
- `X-MockLogistics-Event-Type`
- `X-MockLogistics-Timestamp`
- `X-MockLogistics-Signature`

签名原文为 `timestamp + "." + raw_request_body`，默认允许 5 分钟时间偏差。`data.merchant_order_no` 是业务发货单号，`data.shipment_no` 是物流平台发货单号。

- 相同事件 ID 重复到达时直接返回当前发货单，不重复写轨迹。
- 同一发货单的 sequence 冲突但事件 ID 不同时，整笔回调拒绝。
- 乱序到达的旧事件可以写入轨迹，但不会回退当前状态或清空签收时间。
- 签收事件与子订单、父订单完成在同一事务内提交。
- 业务侧只消费签名回调；轨迹推进节奏由模拟物流平台的自动 worker 控制。

## 数据与依赖

- `shipments` 保存发货单、买卖双方归属、物流平台单号、运单号、承运商、当前状态、收寄信息与时间戳。
- `shipment_events` 保存按发货单和 sequence 唯一的轨迹。
- 两张表属于主业务库，与 `mock_shipments` 等模拟平台表没有外键关系。
- 业务物流与模拟物流不互相 import；业务 Gateway 复用中立的 `internal/http/sessionclient`，签名算法位于 `internal/webhooksig`。

配置见 `backend/.env.example`，HTTP 契约以 controller 与生成的 API 文档为准。
