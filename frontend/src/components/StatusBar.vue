<script setup>
import { computed } from 'vue'
import { usePlayer } from '../composables/usePlayer'
import { useSettings } from '../composables/useSettings'

defineProps({
  eqOpen: { type: Boolean, default: false },
  lyricsOpen: { type: Boolean, default: false },
  sidebarOpen: { type: Boolean, default: true },
})
const emit = defineEmits(['toggle-eq', 'toggle-lyrics', 'toggle-sidebar', 'open-lyrics'])

const { state, currentTrack, cycleRepeat, toggleShuffle } = usePlayer()
const { t } = useSettings()

// The two mode buttons mirror PlayerMain's shuffle / repeat buttons 1:1 and
// share the same player.* labels: shuffle is an on/off toggle (随机播放 /
// 顺序播放), repeat is a 3-state cycle (循环关闭 / 列表循环 / 单曲循环).
const shuffleLabel = computed(() => (state.shuffle ? t('player.shuffleOn') : t('player.shuffleOff')))

const repeatLabel = computed(() => {
  if (state.repeat === 'one') return t('player.repeatOne')
  if (state.repeat === 'all') return t('player.repeatAll')
  return t('player.repeatOff')
})

const statusText = computed(() => {
  const track = currentTrack.value
  if (!track) return t('status.ready')
  const verb = state.isPlaying ? t('status.playing') : t('status.paused')
  return `${verb} · ${track.format}${track.album ? ' · ' + track.album : ''}`
})

// The line currently being sung, shown as a single-line ticker in the center of
// the status bar. Empty before the first timestamp or when the track has no
// lyrics, in which case the ticker is hidden entirely.
const currentLyric = computed(() => {
  const i = state.lyricIndex
  if (i < 0 || !state.lyrics.length) return ''
  return state.lyrics[i]?.text || ''
})
</script>

<template>
  <footer class="bottom-bar">
    <div class="bottom-left">
      <button class="mode-btn playlist-toggle" :class="{ active: sidebarOpen }" :title="t('status.playlist')" @click="emit('toggle-sidebar')">
        <svg viewBox="0 0 24 24" width="16" height="16">
          <path d="M3 6h10M3 10h10M3 14h7M3 18h7" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          <path d="M17 14v7l5-3.5L17 14z" fill="currentColor" />
        </svg>
      </button>
      <span class="ctrl-sep"></span>
      <div class="status-indicator" :class="{ idle: !state.isPlaying }"></div>
      <span class="status-text">{{ statusText }}</span>
    </div>
    <div class="bottom-center">
      <Transition name="lyric-fade" mode="out-in">
        <button
          v-if="currentLyric"
          :key="currentLyric"
          class="lyric-ticker"
          :title="t('status.lyricTickerTitle')"
          @click="emit('open-lyrics')"
        >
          {{ currentLyric }}
        </button>
      </Transition>
    </div>
    <div class="bottom-right">
      <button class="mode-btn" :class="{ active: state.shuffle }" @click="toggleShuffle">{{ shuffleLabel }}</button>
      <button class="mode-btn" :class="{ active: state.repeat !== 'off' }" @click="cycleRepeat">{{ repeatLabel }}</button>
      <button class="mode-btn" :class="{ active: eqOpen }" @click="emit('toggle-eq')">{{ t('status.eq') }}</button>
      <button class="mode-btn" :class="{ active: lyricsOpen }" @click="emit('toggle-lyrics')">{{ t('status.lyrics') }}</button>
    </div>
  </footer>
</template>

<style scoped>
.bottom-bar {
  height: 48px;
  background: var(--surface);
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px 0 12px;
  flex-shrink: 0;
  transition: background 0.4s ease;
}

.bottom-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  max-width: 40%;
}

.status-indicator {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--primary);
  animation: pulse 2s ease-in-out infinite;
  flex-shrink: 0;
}

.status-indicator.idle {
  animation: none;
  background: var(--text-tertiary);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.status-text {
  font-size: 11px;
  color: var(--text-tertiary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bottom-right {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.bottom-center {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8px;
}

.lyric-ticker {
  max-width: 100%;
  font-size: 12px;
  color: var(--text-secondary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.15s, background 0.15s;
}

.lyric-ticker:hover {
  color: var(--primary);
  background: color-mix(in srgb, var(--seed-fg) 8%, transparent);
}

.lyric-fade-enter-active,
.lyric-fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.lyric-fade-enter-from {
  opacity: 0;
  transform: translateY(5px);
}

.lyric-fade-leave-to {
  opacity: 0;
  transform: translateY(-5px);
}

.mode-btn {
  font-size: 11px;
  color: var(--text-tertiary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: color 0.15s, background 0.15s;
}

.mode-btn:hover {
  color: var(--fg);
  background: color-mix(in srgb, var(--seed-fg) 8%, transparent);
}

.mode-btn.active {
  color: var(--primary);
}

.playlist-toggle {
  padding: 4px 6px;
  margin-right: -2px;
}

.playlist-toggle svg {
  display: block;
}

.ctrl-sep {
  width: 1px;
  height: 16px;
  background: var(--border);
  margin: 0 2px;
}
</style>
