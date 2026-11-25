# API 限流

主业务 API 使用 Redis 滑动窗口限流。未认证请求按来源 IP 限流，已认证请求按用户或商户主体限流；Mock 平台不在本策略范围内。

## 四层策略

- API 总量层：全站所有 `/api/` 请求共享一个滑动窗口，不区分 HTTP 方法、路由、用户或商户。该层在认证前执行，因此公开接口、认证接口、业务接口和 Webhook 都会计数。
- 未认证层：登录、注册、签名 Webhook、未知 `/api/` 路径、CORS 预检，以及认证失败的请求按来源 IP 共享一个滑动窗口。Webhook 不豁免该层。
- API 层：按主体角色、主体 ID、HTTP 方法和 Gin 路由模板分别计数，例如 `GET /api/v1/orders/:orderID` 与 `GET /api/v1/feed` 各自使用独立窗口。
- 主体层：按主体角色和 ID 汇总全部请求，用户与商户使用不同命名空间。
- 公开路由和 CORS 预检的 API 总量层与未认证层在单个 Redis Lua 脚本内原子执行，只产生一次 Redis 往返。
- 受保护路由在认证前单独执行 API 总量层；认证成功后，API 路由层和主体层在另一个 Lua 脚本内原子执行。
- Lua 脚本通过 `redis.NewScript().Run()` 执行，go-redis 自动使用并缓存 `EVALSHA`，应用不维护脚本 SHA。
- 超限返回 `429 Too Many Requests` 和 `Retry-After`。`X-RateLimit-Api-Total-*`、`X-RateLimit-Unauthenticated-*`、`X-RateLimit-Api-*`、`X-RateLimit-Subject-*` 分别暴露各层的上限及剩余次数，`X-RateLimit-Policy` 标识触发限制的层。

## 配置

Redis 连接为必需配置。启动时会执行 `Ping`，连接失败则进程拒绝启动。

```text
REDIS_URL=redis://localhost:6379/0
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_API_TOTAL_LIMIT=3000
RATE_LIMIT_UNAUTHENTICATED_LIMIT=60
RATE_LIMIT_API_LIMIT=60
RATE_LIMIT_SUBJECT_LIMIT=300
TRUSTED_PROXIES=
```

`RATE_LIMIT_WINDOW` 使用 Go duration 格式，默认 `1m`，最短为 `1ms`。`RATE_LIMIT_API_TOTAL_LIMIT` 是全站 `/api/` 请求总量上限，默认 `3000`。`RATE_LIMIT_UNAUTHENTICATED_LIMIT` 是单个来源 IP 的未认证请求上限，默认 `60`。另外两个值分别是每个主体在每个 API 路由上的上限和每个主体的总量上限。所有限制值必须配置为正整数。

`TRUSTED_PROXIES` 为空时直接使用 TCP `RemoteAddr` 作为来源 IP。只有部署在可信反向代理后时，才填写用逗号分隔的代理 IP 或 CIDR；Gin 只在这些代理转发的请求上读取 `X-Forwarded-For`，避免客户端伪造来源地址。

本地可通过 Docker 启动 Redis：

```powershell
docker run --name pinshaoshao-redis -p 6379:6379 -d redis:8-alpine
$env:REDIS_URL = "redis://localhost:6379/0"
```

Redis 在请求期间不可用时，限流器记录 `slog` 错误并放行请求，避免缓存基础设施故障扩大为整站不可用。启动阶段的连接和 `Ping` 仍然失败即退出。
