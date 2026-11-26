# API 限流

主业务 API 使用 Redis 滑动窗口限流。Mock 平台不在本策略范围内。

## 四层策略

- API 总量层：所有 `/api/` 请求共享一个窗口，不区分方法、路由或主体；在认证前执行，因此公开接口、登录注册、Webhook 和未知路径都会计数。
- 未认证层：登录、注册、签名 Webhook、未知 `/api/` 路径、CORS 预检和认证失败请求按来源 IP 共享一个窗口；Webhook 不豁免。
- API 层：按主体角色、主体 ID、HTTP 方法和 Gin 路由模板分别计数。
- 主体层：按主体角色和 ID 汇总全部请求；用户与商户使用不同命名空间。

公开路由和 CORS 预检的 API 总量层与未认证层在一个 Lua 脚本内原子执行。受保护路由先单独执行 API 总量层，认证成功后再原子执行 API 路由层与主体层。

超限返回 `429 Too Many Requests` 与 `Retry-After`。`X-RateLimit-Api-Total-*`、`X-RateLimit-Unauthenticated-*`、`X-RateLimit-Api-*`、`X-RateLimit-Subject-*` 暴露各层上限与剩余次数，`X-RateLimit-Policy` 标识触发限制的层。

## 来源 IP 与故障行为

`TRUSTED_PROXIES` 为空时直接使用 TCP `RemoteAddr`。只有部署在可信反向代理后时，才填写代理 IP 或 CIDR；Gin 只在这些代理转发的请求上读取 `X-Forwarded-For`，避免客户端伪造来源地址。

Redis 连接是启动必需项，连接或 `Ping` 失败时进程拒绝启动。请求处理期间 Redis 不可用时，限流器记录 `slog` 错误并放行请求，避免缓存故障扩大为整站不可用。

配置见 `backend/.env.example`。
