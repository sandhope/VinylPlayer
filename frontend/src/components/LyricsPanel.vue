<script setup>
import { watch, ref, nextTick } from 'vue'
import { usePlayer } from '../composables/usePlayer'
import { useSettings } from '../composables/useSettings'

defineEmits(['close'])

const { state, currentTrack } = usePlayer()
const { t } = useSettings()
const listRef = ref(null)

// Keep the active lyric line centered as playback progresses.
watch(
  () => state.lyricIndex,
  async (idx) => {
    if (idx < 0 || !listRef.value) return
    await nextTick()
    const el = listRef.value.querySelector(`[data-line="${idx}"]`)
    if (el) el.scrollIntoView({ block: 'center', behavior: 'smooth' })
  }
)
</script>

<template>
  <div class="panel lyrics-panel">
    <div class="panel-header">
      <div class="panel-title">{{ t('lyrics.title') }}</div>
      <button class="close-btn" :aria-label="t('common.close')" @click="$emit('close')">
        <svg width="14" height="14" viewBox="0 0 12 12">
          <path d="M1 1l10 10M11 1L1 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
      </button>
    </div>

    <div class="now">
      <div class="now-title">{{ currentTrack?.title || t('lyrics.notPlaying') }}</div>
      <div class="now-artist">{{ currentTrack?.artist || '' }}</div>
    </div>

    <div ref="listRef" class="lyrics-list">
      <template v-if="state.lyrics.length">
        <p
          v-for="(line, i) in state.lyrics"
          :key="i"
          :data-line="i"
          class="lyric-line"
          :class="{ active: i === state.lyricIndex }"
        >
          {{ line.text }}
        </p>
      </template>
      <div v-else class="no-lyrics">
        <svg viewBox="0 0 24 24" width="36" height="36">
          <path d="M9 18V5l12-2v13" stroke="currentColor" stroke-width="1.6" fill="none" stroke-linecap="round"
            stroke-linejoin="round" />
          <circle cx="6" cy="18" r="3" stroke="currentColor" stroke-width="1.6" fill="none" />
          <circle cx="18" cy="16" r="3" stroke="currentColor" stroke-width="1.6" fill="none" />
        </svg>
        <p>{{ t('lyrics.empty') }}</p>
        <span>{{ t('lyrics.emptyHint') }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 340px;
  background: color-mix(in srgb, var(--surface) 96%, transparent);
  backdrop-filter: blur(8px);
  border-left: 1px solid var(--border);
  z-index: 20;
  display: flex;
  flex-direction: column;
  padding: 20px;
  box-shadow: -8px 0 32px color-mix(in srgb, var(--shadow-color) 40%, transparent);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--fg);
}

.close-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 6px;
}

.close-btn:hover {
  background: color-mix(in srgb, var(--seed-fg) 10%, transparent);
  color: var(--fg);
}

.now {
  padding-bottom: 14px;
  margin-bottom: 8px;
  border-bottom: 1px solid var(--border);
}

.now-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.now-artist {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 3px;
}

.lyrics-list {
  flex: 1;
  overflow-y: auto;
  padding: 30% 4px;
  text-align: center;
  scroll-behavior: smooth;
}

.lyric-line {
  font-size: 14px;
  line-height: 1.5;
  color: var(--text-tertiary);
  padding: 8px 0;
  transition: color 0.25s, transform 0.25s;
}

.lyric-line.active {
  color: var(--primary);
  font-size: 16px;
  font-weight: 600;
  transform: scale(1.04);
}

.no-lyrics {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--text-tertiary);
}

.no-lyrics p {
  font-size: 14px;
  color: var(--text-secondary);
}

.no-lyrics span {
  font-size: 11px;
  max-width: 200px;
}
</style>
