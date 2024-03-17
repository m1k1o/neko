import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 3001,
    proxy: process.env.NEKO_HOST ? {
      '/api': {
        target: 'http://' + process.env.NEKO_HOST + ':' + process.env.NEKO_PORT + '/',
        changeOrigin: true,
        ws: true
      }
    } : undefined
  }
})
