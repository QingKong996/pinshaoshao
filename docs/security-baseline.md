# 安全基线

本文记录浏览器安全响应头、依赖漏洞扫描门禁和部署侧配合事项。

## 前端 CSP

Vite 在生产构建时向 `index.html` 注入 `Content-Security-Policy` meta，策略默认只允许同源资源。若 `VITE_API_BASE_URL` 是绝对地址，构建会将其 origin 加入 `connect-src`。

`frontend/public/_headers` 会随构建复制到 `dist`。支持 `_headers` 的静态托管平台应直接使用；其他平台需要在反向代理或 CDN 上配置等价响应头。

CSP meta 无法可靠执行 `frame-ancestors`，因此部署侧仍需发送 CSP 响应头或 `X-Frame-Options`。

## HSTS

Go API 的所有响应都包含：

```text
Strict-Transport-Security: max-age=31536000
```

HSTS 只在 HTTPS 响应上生效。前端静态站点与 API 不在同一 origin 时，两个 origin 都必须由 HTTPS 端点发送该响应头。本地开发使用 HTTP，不触发 HSTS。

## 依赖漏洞扫描

CI 在现有测试通过后执行：

```text
cd backend && go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...
cd frontend && npm audit --audit-level=high
```

Go 工具链版本以 `backend/go.mod` 为准，当前为 `1.26.11`。npm 门禁阻断 high 和 critical；moderate 仍会输出，但当前集中在 Storybook、shadcn 等开发工具链，待上游修复后再收紧阈值。
