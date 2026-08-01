import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/auth': 'http://localhost:8080',
      '/webhooks': 'http://localhost:8080',
      '/api/v1/deployments': {
        target: 'http://localhost:8080',
        ws: true,
        changeOrigin: true,
      },
      '/api': 'http://localhost:8080',
    },
  },
})
