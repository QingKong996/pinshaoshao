import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv, type Plugin } from 'vite'
import * as path from 'node:path'

function contentSecurityPolicy(apiBaseURL: string): string {
  const connectSources = ["'self'"]
  if (apiBaseURL) {
    try {
      const url = new URL(apiBaseURL)
      if (url.protocol === 'http:' || url.protocol === 'https:') {
        connectSources.push(url.origin)
      }
    } catch {
      // Relative API base URLs use the page origin.
    }
  }

  return [
    "default-src 'self'",
    "script-src 'self'",
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data: blob:",
    "font-src 'self'",
    `connect-src ${connectSources.join(' ')}`,
    "object-src 'none'",
    "base-uri 'none'",
    "form-action 'self'",
    "frame-ancestors 'none'",
  ].join('; ')
}

function securityHeaders(policy: string): Record<string, string> {
  return {
    'Content-Security-Policy': policy,
    'Strict-Transport-Security': 'max-age=31536000',
    'Permissions-Policy': 'camera=(), geolocation=(), microphone=()',
    'Referrer-Policy': 'no-referrer',
    'X-Content-Type-Options': 'nosniff',
    'X-Frame-Options': 'DENY',
  }
}

function contentSecurityPolicyPlugin(policy: string): Plugin {
  return {
    name: 'security-headers',
    apply: 'build',
    transformIndexHtml(html) {
      return {
        html,
        tags: [
          {
            tag: 'meta',
            attrs: {
              'http-equiv': 'Content-Security-Policy',
              content: policy,
            },
            injectTo: 'head-prepend',
          },
        ],
      }
    },
  }
}

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, import.meta.dirname, '')
  const policy = contentSecurityPolicy(env.VITE_API_BASE_URL ?? '')

  return {
    plugins: [
      react(),
      tailwindcss(),
      contentSecurityPolicyPlugin(policy),
    ],
    resolve: {
      alias: {
        '@': path.resolve(import.meta.dirname, './src'),
      },
    },
    server: {
      host: '127.0.0.1',
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
        '/mockpay/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
        '/mockpay/checkout': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
        '/mocklogistics/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
    preview: {
      headers: securityHeaders(policy),
    },
  }
})
