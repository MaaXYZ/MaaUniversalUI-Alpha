import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    {
      name: 'wails-static-404-handler',
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          const url = req.url || ''

          if (url.startsWith('/static/') || url.startsWith('/resource/')) {
            res.statusCode = 404
            res.end()
            return
          }

          next()
        })
      },
    },
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@wails': resolve(__dirname, 'wailsjs'),
    },
  },
})
