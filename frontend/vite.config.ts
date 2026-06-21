/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// Shared browser baseline. Without this, the dev server transforms with
// esbuild's default `esnext` target and ships modern syntax (optional chaining,
// nullish coalescing, optional catch binding) raw — which older Safari/WebKit
// can't parse, even though `vite build` down-levels and passes. Pinning the same
// target for dev transform, dep pre-bundling, and the production build keeps all
// three consistent and broadly compatible.
const browserTarget = ['es2019', 'safari12', 'chrome80', 'firefox78', 'edge88']

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  esbuild: {
    target: browserTarget,
  },
  optimizeDeps: {
    esbuildOptions: { target: browserTarget },
  },
  build: {
    target: browserTarget,
  },
  server: {
    port: 5173,
    host: true, // expose in Docker (pnpm dev --host)
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test/setup.ts',
    css: true,
  },
})
