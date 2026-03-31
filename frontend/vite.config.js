import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
    assetsDir: 'sqlite-web'
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8903',
        changeOrigin: true
      }
    }
  }
})
