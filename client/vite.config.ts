import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { version } from './package.json'

// vendor code in a few chunks of its own, by package, so an app change does not make browsers refetch
// React and friends (and the build output shows what is big instead of one 700 kB chunk); the rest of
// node_modules is `vendor`, the app itself stays in the entry chunk
const vendorChunks: [string, RegExp][] = [
  ['react', /^(react|react-dom|scheduler)$/],
  [
    'radix',
    /^(@radix-ui\/|@floating-ui\/|aria-hidden|react-remove-scroll|react-style-singleton|use-callback-ref|use-sidecar|get-nonce|tslib)/,
  ],
  [
    'markdown',
    /^(react-markdown|remark-|rehype-|micromark|mdast-|hast-|unified|unist-|vfile|estree-|decode-named-character-reference|@ungap\/structured-clone|property-information|space-separated-tokens|comma-separated-tokens|style-to-|inline-style-parser|html-url-attributes|markdown-table|longest-streak|trim-lines|escape-string-regexp|is-plain-obj|bail|trough|zwitch|ccount|devlop|extend)/,
  ],
  ['i18n', /^(i18next|react-i18next)$/],
  ['lucide', /^lucide-react$/],
]
const vendorChunk = (id: string) => {
  const dep = id.split('/node_modules/').pop()
  if (dep === id) return // not a dependency: app code, core/, a virtual module
  const pkg = /^(@[^/]+\/)?[^/]+/.exec(dep)![0]
  return vendorChunks.find(([, re]) => re.test(pkg))?.[0] ?? 'vendor'
}

// dev: NEKO_URL=http://host:port npm run dev  (proxies /api incl. websocket)
export default defineConfig({
  plugins: [react(), tailwindcss()],
  base: './',
  // shown in the About dialog
  define: { __APP_VERSION__: JSON.stringify(version) },
  resolve: {
    // the core is consumed from source (alias below), so its zustand must be this app's copy, not core/node_modules'
    dedupe: ['zustand'],
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      // the GUI consumes the core package from source; consumers get the built dist/ (core/package.json)
      '@m1k1o/neko': fileURLToPath(new URL('./core/src/index.ts', import.meta.url)),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: vendorChunk,
        // the lazily loaded locale files (src/i18n) by language and namespace, not 15 `common-<hash>.js`
        chunkFileNames: (chunk) => {
          const locale = /\/locales\/(\w+)\/(\w+)\.json$/.exec(chunk.facadeModuleId ?? '')
          return locale ? `assets/locales/${locale[1]}-${locale[2]}-[hash].js` : 'assets/[name]-[hash].js'
        },
      },
    },
  },
  server: {
    port: 3001,
    proxy: process.env.NEKO_URL
      ? { '/api': { target: process.env.NEKO_URL, changeOrigin: true, ws: true } }
      : undefined,
  },
})
