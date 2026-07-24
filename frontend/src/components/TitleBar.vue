<script setup>
import { WindowMinimise, WindowToggleMaximise, Quit } from '../../wailsjs/go/main/App'
</script>

<template>
  <header class="title-bar" style="--wails-draggable:drag">
    <div class="title-bar-left">
      <svg class="app-icon" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="var(--primary)" stroke-width="2" />
        <circle cx="12" cy="12" r="3" fill="var(--primary)" />
        <path d="M12 2a10 10 0 0 1 0 20" stroke="var(--accent)" stroke-width="1" opacity="0.5" />
      </svg>
      <span class="title-bar-text">Vinyl Player · 本地音乐播放器</span>
    </div>
    <div class="window-controls" style="--wails-draggable:no-drag">
      <button class="win-btn" aria-label="最小化" @click="WindowMinimise">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <rect y="5" width="12" height="1.5" fill="currentColor" />
        </svg>
      </button>
      <button class="win-btn" aria-label="最大化" @click="WindowToggleMaximise">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <rect x="1" y="1" width="10" height="10" rx="1.5" stroke="currentColor" stroke-width="1.5" fill="none" />
        </svg>
      </button>
      <button class="win-btn close" aria-label="关闭" @click="Quit">
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
