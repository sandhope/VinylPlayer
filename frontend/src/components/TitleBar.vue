<script setup>
import { ref, onMounted } from 'vue'
import { WindowMinimise, HideToTray } from '../../wailsjs/go/main/App'
import { useSettings } from '../composables/useSettings'

defineEmits(['open-settings', 'open-about'])
const { t } = useSettings()

// 跟踪窗口最大化状态，切换「最大化 / 还原」图标（与 Pixel 一致）
const isMax = ref(false)
function syncMax() {
  try {
    window.runtime?.WindowIsMaximised?.().then((v) => {
      isMax.value = !!v
    })
  } catch (e) {
    /* 浏览器预览时 window.runtime 不存在，忽略 */
  }
}
function toggleMax() {
  window.runtime?.WindowToggleMaximise?.()
  setTimeout(syncMax, 60)
}
onMounted(() => syncMax())
</script>

<template>
  <header class="title-bar" style="--wails-draggable:drag">
    <div class="title-bar-left">
      <svg class="app-icon" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="var(--primary)" stroke-width="2" />
        <circle cx="12" cy="12" r="3" fill="var(--primary)" />
        <path d="M12 2a10 10 0 0 1 0 20" stroke="var(--accent)" stroke-width="1" opacity="0.5" />
      </svg>
      <span class="title-bar-text">Vinyl Player · {{ t('app.subtitle') }}</span>
    </div>
    <div class="window-controls" style="--wails-draggable:no-drag">
      <button class="win-btn" :aria-label="t('titlebar.about')" :title="t('titlebar.about')" @click="$emit('open-about')">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.6" />
          <path d="M12 11v5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
          <circle cx="12" cy="8" r="1" fill="currentColor" />
        </svg>
      </button>
      <button class="win-btn" :aria-label="t('titlebar.settings')" :title="t('titlebar.settings')" @click="$emit('open-settings')">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.8" />
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"
            stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
      <button class="win-btn" :aria-label="t('titlebar.minimize')" @click="WindowMinimise">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <rect y="5" width="12" height="1.5" fill="currentColor" />
        </svg>
      </button>
      <button class="win-btn" :aria-label="t('titlebar.maximize')" @click="toggleMax">
        <svg v-if="!isMax" width="12" height="12" viewBox="0 0 12 12">
          <rect x="1" y="1" width="10" height="10" rx="1.5" stroke="currentColor" stroke-width="1.5" fill="none" />
        </svg>
        <svg v-else width="12" height="12" viewBox="0 0 12 12">
          <rect x="3" y="1" width="8" height="8" rx="1.2" stroke="currentColor" stroke-width="1.4" fill="none" />
          <rect x="1" y="3" width="8" height="8" rx="1.2" stroke="currentColor" stroke-width="1.4" fill="var(--surface-sunken)" />
        </svg>
      </button>
      <button class="win-btn close" :aria-label="t('titlebar.close')" @click="HideToTray">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <path d="M1 1l10 10M11 1L1 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
      </button>
    </div>
  </header>
</template>

<style scoped>
.title-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 36px;
  padding: 0 8px 0 12px;
  background: var(--surface-sunken);
  border-bottom: 1px solid var(--border);
  user-select: none;
  flex-shrink: 0;
}

.title-bar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.app-icon {
  width: 18px;
  height: 18px;
}

.title-bar-text {
  font-size: 12px;
  color: var(--text-secondary);
  letter-spacing: 0.3px;
}

.window-controls {
  display: flex;
  gap: 2px;
}

.win-btn {
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
  transition: background 0.15s;
}

.win-btn:hover {
  background: color-mix(in srgb, var(--seed-fg) 10%, transparent);
}

.win-btn.close:hover {
  background: #e81123;
  color: #fff;
}
</style>
