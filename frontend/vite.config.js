import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  base: './',
  server: {
    port: 34115
  },
  root: 'src',  // Set src as root
  build: {
    outDir: '../dist',
    rollupOptions: {
      input: path.resolve(__dirname, 'src/index.html')  // Explicit entry
    }
  }
})
