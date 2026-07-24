<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import { useSettings } from '../composables/useSettings'

const emit = defineEmits(['close'])
const { settings, t, setLocale } = useSettings()

const locales = [
  { id: 'zh', label: '中文' },
  { id: 'en', label: 'English' },
]

function onKeydown(e) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="settings-overlay" @click.self="emit('close')">
    <div class="settings-card" role="dialog" aria-modal="true">
      <div class="settings-header">
        <div class="settings-title">{{ t('settings.title') }}</div>
        <button class="close-btn" :aria-label="t('common.close')" @click="emit('close')">
          <svg width="14" height="14" viewBox="0 0 12 12">
            <path d="M1 1l10 10M11 1L1 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          </svg>
        </button>
      </div>

      <div class="settings-body">
        <div class="setting-row">
          <div class="setting-label">
            <div class="setting-name">{{ t('settings.language') }}</div>
            <div class="setting-desc">{{ t('settings.languageDesc') }}</div>
          </div>
          <div class="segmented">
            <button
              v-for="l in locales"
              :key="l.id"
              class="seg-btn"
              :class="{ active: settings.locale === l.id }"
              @click="setLocale(l.id)"
            >
              {{ l.label }}
            </button>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-label">
            <div class="setting-name">{{ t('settings.rememberProgress') }}</div>
            <div class="setting-desc">{{ t('settings.rememberProgressDesc') }}</div>
          </div>
          <label class="switch">
            <input type="checkbox" v-model="settings.rememberProgress" />
            <span class="slider-toggle"></span>
          </label>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-overlay {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, #000 40%, transparent);
  backdrop-filter: blur(2px);
}

.settings-card {
  width: 420px;
  max-width: calc(100vw - 48px);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: 0 16px 48px color-mix(in srgb, var(--shadow-color) 55%, transparent);
  overflow: hidden;
}

.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 8px 0 18px;
  background: var(--surface-sunken);
  border-bottom: 1px solid var(--border);
}

.settings-title {
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

.settings-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 22px 20px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 0;
  border-bottom: 1px solid var(--border);
}

.setting-row:last-child {
  border-bottom: none;
}

.setting-label {
  min-width: 0;
}

.setting-name {
  font-size: 13.5px;
  font-weight: 500;
  color: var(--fg);
}

.setting-desc {
  font-size: 11.5px;
  color: var(--text-tertiary);
  margin-top: 3px;
}

.segmented {
  display: flex;
  gap: 2px;
  padding: 3px;
  background: color-mix(in srgb, var(--seed-fg) 6%, transparent);
  border-radius: calc(var(--radius) * 0.6);
  flex-shrink: 0;
}

.seg-btn {
  padding: 5px 14px;
  font-size: 12px;
  color: var(--text-secondary);
  background: transparent;
  border: none;
  border-radius: calc(var(--radius) * 0.45);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.seg-btn:hover {
  color: var(--fg);
}

.seg-btn.active {
  color: var(--fg);
  background: var(--surface);
  box-shadow: 0 1px 3px color-mix(in srgb, var(--shadow-color) 30%, transparent);
}

/* Toggle switch (matches the EQ panel switch) */
.switch {
  position: relative;
  display: inline-block;
  width: 34px;
  height: 18px;
  flex-shrink: 0;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider-toggle {
  position: absolute;
  inset: 0;
  background: var(--progress-track);
  border-radius: 10px;
  transition: background 0.2s;
  cursor: pointer;
}

.slider-toggle::before {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  left: 2px;
  top: 2px;
  background: var(--fg);
  border-radius: 50%;
  transition: transform 0.2s;
}

.switch input:checked + .slider-toggle {
  background: var(--primary);
}

.switch input:checked + .slider-toggle::before {
  transform: translateX(16px);
  background: var(--bg);
}
</style>
