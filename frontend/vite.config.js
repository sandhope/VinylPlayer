import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

// Single source of truth for the app version: read info.productVersion from
// wails.json (which also drives the Windows executable metadata) and expose it
// to the frontend as __APP_VERSION__ so the About page never drifts from it.
let appVersion = '0.0.0'
try {
  const wailsConfig = JSON.parse(
    readFileSync(fileURLToPath(new URL('../wails.json', import.meta.url)), 'utf-8')
  )
  appVersion = (wailsConfig.info && wailsConfig.info.productVersion) || appVersion
} catch (e) {
  /* fall back to default */
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
})
