<script setup>
import { computed } from 'vue'
import { usePlayer } from '../composables/usePlayer'

defineProps({
  eqOpen: { type: Boolean, default: false },
  lyricsOpen: { type: Boolean, default: false },
})
const emit = defineEmits(['toggle-eq', 'toggle-lyrics', 'open-lyrics'])

const { state, currentTrack } = usePlayer()

const playMode = computed(() => {
  if (state.shuffle) return '随机播放'
  if (state.repeat === 'one') return '单曲循环'
  if (state.repeat === 'all') return '列表循环'
  return '顺序播放'
})

function cyclePlayMode() {
  // 顺序 -> 列表循环 -> 单曲循环 -> 随机 -> 顺序
  if (state.shuffle) {
    state.shuffle = false
    state.repeat = 'off'
  } else if (state.repeat === 'off') {
    state.repeat = 'all'
  } else if (state.repeat === 'all') {
    state.repeat = 'one'
  } else {
    state.repeat = 'off'
    state.shuffle = true
  }
}

const statusText = computed(() => {
  const t = currentTrack.value
  if (!t) return '就绪 · 等待播放'
  const verb = state.isPlaying ? '正在播放' : '已暂停'
  return `${verb} · ${t.format}${t.album ? ' · ' + t.album : ''}`
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
      <div class="status-indicator" :class="{ idle: !state.isPlaying }"></div>
      <span class="status-text">{{ statusText }}</span>
    </div>
    <div class="bottom-center">
      <Transition name="lyric-fade" mode="out-in">
        <button
          v-if="currentLyric"
          :key="currentLyric"
          class="lyric-ticker"
          title="点击查看完整歌词"
          @click="emit('open-lyrics')"
        >
          {{ currentLyric }}
        </button>
      </Transition>
    </div>
    <div class="bottom-right">
      <button class="mode-btn" @click="cyclePlayMode">{{ playMode }}</button>
      <button class="mode-btn" :class="{ active: eqOpen }" @click="emit('toggle-eq')">均衡器</button>
      <button class="mode-btn" :class="{ active: lyricsOpen }" @click="emit('toggle-lyrics')">歌词</button>
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
  padding: 0 20px;
  flex-shrink: 0;
  transition: background 0.4s ease;
}

.bottom-left {
  display: flex;
  align-items: center;
  gap: 12px;
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
</style>
