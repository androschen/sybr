import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    port: 34115,
    strictPort: true, // Fail if port is already in use
    host: true, // Listen on all addresses
  },
  // Ensure base path is correct
  base: '/',
})

