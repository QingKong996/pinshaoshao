# 业务支付

`internal/feature/payment` 是商城侧的支付 feature。它不实现收银台；渠道能力由同一进程中的 `internal/mock/payment` 通过 `/mockpay` 提供。双方只通过 HTTP API 和签名回调通信，不互相 import。模拟支付平台见 [mock-payment.md](mock-payment.md)。

## 职责边界

- `payment` 拥有 `payments` 表，每个父订单最多一张业务支付单。
- `payment_no` 既是业务支付单号，也是向渠道创建支付会话的幂等键。
- 创建业务支付单只落本地记录；发起支付时才创建或复用渠道支付会话。
- 支付回调验签、核对金额与渠道流水，并在一个事务内推进支付单、父订单和子订单。
- 主动关单先完成本地订单取消；渠道关单经过本地 outbox 异步投递。
- 0 元订单在创建时直接置为 `paid`，子订单置为 `awaiting_shipment`；这类订单不创建支付单，也不调用第三方支付。

## 状态与一致性

业务支付单状态：

```text
pending ──payment callback──> succeeded
   │
   └──close─────────────────> closed ──late signed callback──> succeeded
```

- 创建支付单时锁定父订单；只有未过期且仍为 `pending_payment` 的订单可以创建。
- 同一订单重复创建支付单返回已有记录，不会生成第二张业务支付单。
- 发起支付要求订单剩余有效期不少于 60 秒；重复发起使用同一 `payment_no` 复用渠道支付单。
- 支付成功回调先锁父订单，再锁支付单。`pending_payment` 订单推进为 `paid`，其子订单推进为 `awaiting_shipment`。
- 相同渠道流水的重复回调直接成功；金额或渠道流水不一致时整笔回滚。
- 已关闭的支付单仍可被迟到的签名回调推进为 `succeeded`，因为渠道可能先完成扣款再处理关单。
- 订单已取消或已过期后收到支付成功回调时，支付单记录成功与渠道流水，但订单不会恢复，库存也不会重新冻结。
- 过期订单的回调先触发订单过期取消，再重试结算；最终结果是订单保持 `cancelled`，支付单记录 `succeeded`。

## 关单与 Outbox

主动关单和订单过期共用同一条取消链路：

1. 在同一业务事务中锁定订单与支付单。
2. 取消父订单、取消子订单并释放冻结库存。
3. 将支付单置为 `closed`。
4. 如果已经存在渠道支付单，向 `outbox_events` 写入一条 `payment.close_provider` 事件。

渠道关单由后台 worker 异步执行，业务事务不等待渠道响应。outbox 事件使用 `payment.close:<paymentNo>` 作为唯一键；重复取消不会产生重复事件。worker 采用租约领取和至少一次投递，关闭渠道支付单的 handler 必须保持幂等。

worker 启动后立即处理并在之后轮询；失败按退避重试，达到最大次数或遇到永久错误后进入 `dead`。因此本地取消与事件持久化原子提交，渠道关单异步执行；永久失败需要人工处理。

订单另有独立过期 worker：每分钟扫描最多 100 个超过 30 分钟支付期限的订单，并走上述取消链路。

## 回调协议

回调头：

- `X-MockPay-Event-Id`
- `X-MockPay-Event-Type: payment.paid`
- `X-MockPay-Timestamp`
- `X-MockPay-Signature: v1=<hex>`

签名原文为 `timestamp + "." + raw_request_body`。业务侧使用原始请求体验签，再解析 JSON；默认允许 5 分钟时间偏差。`merchant_order_no` 对应业务支付单号，`payment_no` 对应渠道流水号。

## 配置与前端

环境变量见 `backend/.env.example`，数据库迁移见 `backend/migrations`。回调地址必须与 MockPay 的回调白名单完全一致。

前端支付入口位于买家订单详情页。`PAYMENT_RETURN_URL_BASE` 必须与买家实际访问前端时使用的 origin 完全一致，否则支付返回后会因 `localStorage` 的 origin 隔离而丢失登录状态。
