<script setup>
import { ref, onBeforeUnmount } from 'vue'
import { usePlayer } from '../composables/usePlayer'
import { formatTime } from '../composables/format'

defineProps({
  busy: { type: Boolean, default: false },
})
const emit = defineEmits(['add-folder', 'add-files', 'remove-track', 'clear-all'])

const { state, loadIndex } = usePlayer()

function thumbShade(i) {
  return 0.55 + (i % 4) * 0.12
}

// Two-step confirm for clearing the whole list, avoiding an accidental wipe
// without needing a modal dialog.
const confirmingClear = ref(false)
let confirmTimer = null
function onClearClick() {
  if (confirmingClear.value) {
    clearTimeout(confirmTimer)
    confirmingClear.value = false
    emit('clear-all')
    return
  }
  confirmingClear.value = true
  confirmTimer = setTimeout(() => (confirmingClear.value = false), 2500)
}
onBeforeUnmount(() => clearTimeout(confirmTimer))
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <span class="sidebar-title">播放列表</span>
      <div class="header-right">
        <span class="track-count">{{ state.tracks.length }} 首</span>
        <button
          v-if="state.tracks.length"
          class="clear-btn"
          :class="{ confirming: confirmingClear }"
          :title="confirmingClear ? '再次点击确认清空' : '清空播放列表'"
          @click="onClearClick"
        >
          <template v-if="confirmingClear">确认?</template>
          <svg v-else viewBox="0 0 24 24" width="14" height="14">
            <path d="M4 7h16M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2m1 0v12a1 1 0 0 1-1 1H8a1 1 0 0 1-1-1V7"
              stroke="currentColor" stroke-width="1.8" fill="none" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
      </div>
    </div>

    <div class="sidebar-actions">
      <button class="add-btn" :disabled="busy" @click="emit('add-folder')">
        <svg viewBox="0 0 24 24" width="15" height="15">
          <path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"
            stroke="currentColor" stroke-width="1.8" fill="none" stroke-linejoin="round" />
        </svg>
        添加文件夹
      </button>
      <button class="add-btn" :disabled="busy" @click="emit('add-files')">
        <svg viewBox="0 0 24 24" width="15" height="15">
          <path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8l-5-5z"
            stroke="currentColor" stroke-width="1.8" fill="none" stroke-linejoin="round" />
          <path d="M14 3v5h5" stroke="currentColor" stroke-width="1.8" fill="none" stroke-linejoin="round" />
        </svg>
        添加文件
      </button>
    </div>

    <div class="playlist">
      <div
        v-for="(t, i) in state.tracks"
        :key="t.id"
        class="track-item"
        :class="{ active: i === state.currentIndex }"
        @dblclick="loadIndex(i, true)"
        @click="loadIndex(i, true)"
      >
        <span class="track-index">
          <template v-if="i === state.currentIndex && state.isPlaying">♪</template>
          <template v-else>{{ i + 1 }}</template>
        </span>
        <div class="track-thumb">
          <img v-if="t.coverUrl" :src="t.coverUrl" alt="" />
          <svg v-else viewBox="0 0 40 40">
            <rect width="40" height="40" fill="var(--seed-primary)" :opacity="thumbShade(i)" />
            <circle cx="20" cy="20" r="12" fill="var(--seed-accent)" opacity="0.5" />
            <circle cx="20" cy="20" r="4" fill="var(--seed-surface)" />
          </svg>
        </div>
        <div class="track-info">
          <div class="track-name">{{ t.title }}</div>
          <div class="track-artist">{{ t.artist }}</div>
        </div>
        <span class="track-duration">{{ t.duration ? formatTime(t.duration) : t.format }}</span>
        <button
          class="track-remove"
          title="从列表移除"
          @click.stop="emit('remove-track', t.id)"
        >
          <svg viewBox="0 0 24 24" width="14" height="14">
            <path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          </svg>
        </button>
      </div>

      <div v-if="!state.tracks.length" class="empty-hint">
        <p>播放列表为空</p>
        <span>点击上方按钮添加本地音乐</span>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 300px;
  background: var(--surface);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  transition: background 0.4s ease;
}

.sidebar-header {
  padding: 16px 20px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.sidebar-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
  letter-spacing: 0.2px;
}

.track-count {
  font-size: 11px;
  color: var(--text-tertiary);
  background: color-mix(in srgb, var(--seed-fg) 8%, transparent);
  padding: 2px 8px;
  border-radius: 10px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.clear-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 22px;
  padding: 0 7px;
  font-size: 11px;
  color: var(--text-tertiary);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: color 0.15s, background 0.15s, border-color 0.15s;
}

.clear-btn:hover {
  color: var(--danger, #d9534f);
  background: color-mix(in srgb, var(--danger, #d9534f) 12%, transparent);
}

.clear-btn.confirming {
  color: #fff;
  background: var(--danger, #d9534f);
  border-color: var(--danger, #d9534f);
}

.sidebar-actions {
  display: flex;
  gap: 8px;
  padding: 0 16px 12px;
}

.add-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 6px;
  font-size: 12px;
  color: var(--text-secondary);
  background: color-mix(in srgb, var(--seed-fg) 5%, transparent);
  border: 1px solid var(--border);
  border-radius: calc(var(--radius) * 0.6);
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.add-btn:hover:not(:disabled) {
  color: var(--fg);
  background: color-mix(in srgb, var(--seed-fg) 10%, transparent);
  border-color: color-mix(in srgb, var(--seed-primary) 40%, transparent);
}

.add-btn:disabled {
  opacity: 0.5;
  cursor: default;
}

.playlist {
  flex: 1;
  overflow-y: auto;
  padding: 4px 8px;
}

.track-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: calc(var(--radius) * 0.6);
  cursor: pointer;
  transition: background 0.15s;
  position: relative;
}

.track-item:hover {
  background: color-mix(in srgb, var(--seed-fg) 6%, transparent);
}

.track-item.active {
  background: var(--accent-soft);
}

.track-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: var(--primary);
  border-radius: 2px;
}

.track-index {
  width: 20px;
  font-size: 12px;
  color: var(--text-tertiary);
  text-align: center;
  flex-shrink: 0;
}

.track-item.active .track-index {
  color: var(--primary);
}

.track-thumb {
  width: 40px;
  height: 40px;
  border-radius: calc(var(--radius) * 0.5);
  overflow: hidden;
  flex-shrink: 0;
  position: relative;
  background: var(--surface-sunken);
}

.track-thumb svg,
.track-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.track-info {
  flex: 1;
  min-width: 0;
}

.track-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--fg);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.track-artist {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.track-duration {
  font-size: 11px;
  color: var(--text-tertiary);
  flex-shrink: 0;
  font-variant-numeric: tabular-nums;
  transition: opacity 0.15s;
}

.track-remove {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  color: var(--text-tertiary);
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s, color 0.15s, background 0.15s;
}

.track-item:hover .track-remove {
  opacity: 1;
  pointer-events: auto;
}

.track-item:hover .track-duration {
  opacity: 0;
}

.track-remove:hover {
  color: var(--danger, #d9534f);
  background: color-mix(in srgb, var(--danger, #d9534f) 14%, transparent);
}

.empty-hint {
  text-align: center;
  padding: 48px 20px;
  color: var(--text-tertiary);
}

.empty-hint p {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.empty-hint span {
  font-size: 11px;
}
</style>
