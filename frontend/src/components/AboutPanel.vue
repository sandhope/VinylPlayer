<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useSettings } from '../composables/useSettings'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

const emit = defineEmits(['close'])
const { settings, t } = useSettings()

// App version comes from wails.json (info.productVersion), injected at build
// time by vite.config.js so there is a single source of truth.
/* global __APP_VERSION__ */
const APP_VERSION = 'v' + __APP_VERSION__
const PROJECT_URL = 'https://github.com/sandhope/VinylPlayer'
const RELEASES_URL = 'https://github.com/sandhope/VinylPlayer/releases'

// Donation QR images live in frontend/public and are referenced by absolute
// path, so they can be dropped in without touching code. If a file is missing
// the <img> onerror flips a flag and we show a placeholder hint instead.
const alipayBroken = ref(false)
const wechatBroken = ref(false)

// Runtime string paths (not static src) so the SFC compiler leaves them as
// public-dir URLs instead of trying to resolve them as bundled assets.
const alipaySrc = '/donate-alipay.jpg'
const wechatSrc = '/donate-wechat.jpg'

function openProject() {
  BrowserOpenURL(PROJECT_URL)
}
function openReleases() {
  BrowserOpenURL(RELEASES_URL)
}

function onKeydown(e) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="about-overlay" @click.self="emit('close')">
    <div class="about-card" role="dialog" aria-modal="true">
      <div class="about-header">
        <div class="about-title">{{ t('about.title') }}</div>
        <button class="close-btn" :aria-label="t('common.close')" @click="emit('close')">
          <svg width="14" height="14" viewBox="0 0 12 12">
            <path d="M1 1l10 10M11 1L1 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          </svg>
        </button>
      </div>

      <div class="about-body">
        <div class="brand">
          <svg class="brand-icon" viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="10" stroke="var(--primary)" stroke-width="2" />
            <circle cx="12" cy="12" r="3" fill="var(--primary)" />
            <path d="M12 2a10 10 0 0 1 0 20" stroke="var(--accent)" stroke-width="1" opacity="0.5" />
          </svg>
          <div class="brand-text">
            <div class="brand-name">Vinyl Player</div>
            <div class="brand-tagline">{{ t('about.tagline') }}</div>
            <div class="brand-version">{{ t('about.version') }} {{ APP_VERSION }}</div>
          </div>
        </div>

        <div class="link-row" @click="openProject">
          <div class="link-icon">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.58 2 12.25c0 4.53 2.87 8.37 6.84 9.73.5.1.68-.22.68-.49 0-.24-.01-.88-.01-1.73-2.78.62-3.37-1.37-3.37-1.37-.45-1.18-1.11-1.5-1.11-1.5-.91-.64.07-.63.07-.63 1 .07 1.53 1.06 1.53 1.06.89 1.56 2.34 1.11 2.91.85.09-.66.35-1.11.63-1.37-2.22-.26-4.55-1.14-4.55-5.07 0-1.12.39-2.03 1.03-2.75-.1-.26-.45-1.3.1-2.71 0 0 .84-.28 2.75 1.05a9.3 9.3 0 0 1 5 0c1.91-1.33 2.75-1.05 2.75-1.05.55 1.41.2 2.45.1 2.71.64.72 1.03 1.63 1.03 2.75 0 3.94-2.34 4.81-4.57 5.06.36.32.68.94.68 1.9 0 1.37-.01 2.48-.01 2.82 0 .27.18.6.69.49A10.02 10.02 0 0 0 22 12.25C22 6.58 17.52 2 12 2z" />
            </svg>
          </div>
          <div class="link-info">
            <div class="link-name">{{ t('about.projectHome') }}</div>
            <div class="link-desc">{{ t('about.projectHomeDesc') }}</div>
          </div>
          <svg class="link-arrow" viewBox="0 0 24 24" width="16" height="16">
            <path d="M7 17L17 7M17 7H9M17 7v8" stroke="currentColor" stroke-width="1.8" fill="none" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </div>

        <div class="link-row" @click="openReleases">
          <div class="link-icon">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none">
              <path d="M12 3v12m0 0l-4-4m4 4l4-4M5 19h14" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </div>
          <div class="link-info">
            <div class="link-name">{{ t('about.checkUpdate') }}</div>
            <div class="link-desc">{{ t('about.checkUpdateDesc') }}</div>
          </div>
          <svg class="link-arrow" viewBox="0 0 24 24" width="16" height="16">
            <path d="M7 17L17 7M17 7H9M17 7v8" stroke="currentColor" stroke-width="1.8" fill="none" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </div>

        <div v-if="settings.locale === 'zh'" class="donate">
          <div class="donate-title">{{ t('about.donateTitle') }}</div>
          <div class="donate-desc">{{ t('about.donateDesc') }}</div>
          <div class="qr-grid">
            <div class="qr-item">
              <div class="qr-frame">
                <img
                  v-show="!alipayBroken"
                  :src="alipaySrc"
                  :alt="t('about.alipay')"
                  @error="alipayBroken = true"
                />
                <div v-if="alipayBroken" class="qr-missing">{{ t('about.qrMissing', { file: 'donate-alipay.jpg' }) }}</div>
              </div>
              <div class="qr-label alipay">{{ t('about.alipay') }}</div>
            </div>
            <div class="qr-item">
              <div class="qr-frame">
                <img
                  v-show="!wechatBroken"
                  :src="wechatSrc"
                  :alt="t('about.wechat')"
                  @error="wechatBroken = true"
                />
                <div v-if="wechatBroken" class="qr-missing">{{ t('about.qrMissing', { file: 'donate-wechat.jpg' }) }}</div>
              </div>
              <div class="qr-label wechat">{{ t('about.wechat') }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.about-overlay {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, #000 40%, transparent);
  backdrop-filter: blur(2px);
}

.about-card {
  width: 420px;
  max-width: calc(100vw - 48px);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: 0 16px 48px color-mix(in srgb, var(--shadow-color) 55%, transparent);
  overflow: hidden;
}

.about-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 8px 0 18px;
  background: var(--surface-sunken);
  border-bottom: 1px solid var(--border);
}

.about-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
  letter-spacing: 0.3px;
}

.close-btn {
  width: 36px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.15s, color 0.15s;
}

.close-btn:hover {
  background: #e81123;
  color: #fff;
}

.about-body {
  padding: 20px 22px 22px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--border);
}

.brand-icon {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
}

.brand-text {
  min-width: 0;
}

.brand-name {
  font-size: 17px;
  font-weight: 700;
  color: var(--fg);
  letter-spacing: 0.3px;
}

.brand-tagline {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 3px;
}

.brand-version {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
  font-variant-numeric: tabular-nums;
}

.link-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 10px;
  margin: 6px -10px 0;
  border-radius: calc(var(--radius) * 0.6);
  cursor: pointer;
  transition: background 0.15s;
}

.link-row:hover {
  background: color-mix(in srgb, var(--seed-fg) 6%, transparent);
}

.link-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  color: var(--text-secondary);
  background: color-mix(in srgb, var(--seed-fg) 6%, transparent);
  border-radius: calc(var(--radius) * 0.5);
}

.link-info {
  flex: 1;
  min-width: 0;
}

.link-name {
  font-size: 13.5px;
  font-weight: 500;
  color: var(--fg);
}

.link-desc {
  font-size: 11.5px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

.link-arrow {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.link-row:hover .link-arrow {
  color: var(--primary);
}

.donate {
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid var(--border);
}

.donate-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--fg);
  text-align: center;
}

.donate-desc {
  font-size: 11.5px;
  color: var(--text-tertiary);
  text-align: center;
  margin-top: 4px;
}

.qr-grid {
  display: flex;
  gap: 16px;
  margin-top: 14px;
}

.qr-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.qr-frame {
  width: 100%;
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  background: #fff;
  border: 1px solid var(--border);
  border-radius: calc(var(--radius) * 0.6);
  overflow: hidden;
}

.qr-frame img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.qr-missing {
  font-size: 10.5px;
  line-height: 1.5;
  color: #888;
  text-align: center;
  padding: 4px;
}

.qr-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.qr-label.alipay {
  color: #1678ff;
}

.qr-label.wechat {
  color: #2aae67;
}
</style>
