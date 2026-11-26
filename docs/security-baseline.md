# 安全基线

本文只记录浏览器安全响应头和部署侧约束。依赖漏洞扫描门禁见 `.github/workflows/ci.yml`。

## 前端 CSP

Vite 在生产构建时向 `index.html` 注入 `Content-Security-Policy` meta，默认只允许同源资源。若 `VITE_API_BASE_URL` 是绝对地址，构建会将其 origin 加入 `connect-src`。

`frontend/public/_headers` 会随构建复制到 `dist`。支持 `_headers` 的静态托管平台应直接使用；其他平台需要在反向代理或 CDN 上配置等价响应头。

CSP meta 无法可靠执行 `frame-ancestors`，因此部署侧仍需发送 CSP 响应头或 `X-Frame-Options`。

## HSTS

Go API 的所有响应都包含 `Strict-Transport-Security: max-age=31536000`。HSTS 只在 HTTPS 响应上生效；前端静态站点与 API 不在同一 origin 时，两个 origin 都必须由 HTTPS 端点发送该响应头。

本地开发使用 HTTP，不触发 HSTS。
