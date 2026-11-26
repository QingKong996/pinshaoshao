# Mock 数据库隔离

MockPay 与 MockLogistics 使用同一个 PostgreSQL 服务中的两个独立逻辑数据库：

```text
pinshaoshao
pinshaoshao_mockpay
pinshaoshao_mocklogistics
```

这不是三个 PostgreSQL 服务、容器或端口。三个连接分别指向同一个 PostgreSQL 实例中的三个 database catalog，各自拥有连接池、迁移目录、表和 Session。

## 创建数据库

先用数据库管理员账号创建另外两个 database。`CREATE DATABASE` 不能在事务中执行：

```bash
createdb -h localhost -U postgres pinshaoshao_mockpay
createdb -h localhost -U postgres pinshaoshao_mocklogistics
```

## 连接与隔离校验

服务启动必须显式提供三个 DSN，Mock DSN 不会回退到 `DATABASE_URL`：

```text
DATABASE_URL
MOCKPAY_DATABASE_URL
MOCKLOGISTICS_DATABASE_URL
```

启动时会分别执行 `Ping` 和 `SELECT current_database()`。如果任意两个连接落在同一个数据库中，进程直接拒绝启动。

运营账号和 Session 配置见 `backend/.env.example`；密码哈希、Session 轮换与撤销由 `internal/mock/auth` 实现。

## 迁移

三个目录分别执行：

```bash
cd backend

goose -dir ./migrations postgres "$DATABASE_URL" up
goose -dir ./migrations/mockpay postgres "$MOCKPAY_DATABASE_URL" up
goose -dir ./migrations/mocklogistics postgres "$MOCKLOGISTICS_DATABASE_URL" up
```

PowerShell 使用 `$env:DATABASE_URL` 等形式传入环境变量。

主库迁移目录保留历史 Mock 迁移，便于已执行旧迁移的数据库继续升级；随后执行的主库清理迁移会删除这些 Mock 表。旧主库中的 Mock 开发数据不会自动复制到新逻辑库。

## 边界

- MockPay 只连接 `pinshaoshao_mockpay`，拥有 `mock_payment_*` 与 `mock_auth_*` 表。
- MockLogistics 只连接 `pinshaoshao_mocklogistics`，拥有 `mock_shipments`、`mock_logistics_*` 与 `mock_auth_*` 表。
- 主业务只连接 `pinshaoshao`，不读取或写入 Mock 表。
- 业务支付、物流 Gateway 只通过 HTTP 调用 Mock 服务，并复用对应运营账号登录。
