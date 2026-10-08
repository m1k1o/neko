import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// dev: NEKO_URL=http://host:port npm run dev  (proxies /api incl. websocket)
export default defineConfig({
  plugins: [react()],
  base: './',
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      // the GUI consumes the core package from source; consumers get the built dist/ (core/package.json)
      '@m1k1o/neko': fileURLToPath(new URL('./core/src/index.ts', import.meta.url)),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        // legacy styles rely on global variables and @import
        additionalData: '@import "@/design/variables";\n',
        silenceDeprecations: ['import', 'global-builtin', 'color-functions'],
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
