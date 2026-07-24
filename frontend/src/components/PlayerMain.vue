<script setup>
import { computed, ref } from 'vue'
import { usePlayer } from '../composables/usePlayer'
import { useSettings } from '../composables/useSettings'
import { formatTime } from '../composables/format'
import ThemeSwitcher from './ThemeSwitcher.vue'

const {
  state,
  currentTrack,
  togglePlay,
  next,
  prev,
  seekFraction,
  setVolume,
  toggleMute,
  toggleShuffle,
  cycleRepeat,
} = usePlayer()
const { t } = useSettings()

const BAR_COUNT = 48
const bars = Array.from({ length: BAR_COUNT }, (_, i) => i)

const progressPct = computed(() =>
  state.duration ? (state.currentTime / state.duration) * 100 : 0
)
const volumePct = computed(() => (state.muted ? 0 : state.volume * 100))

const title = computed(() => currentTrack.value?.title || t('player.noTrack'))
const artist = computed(() => currentTrack.value?.artist || '—')

function barHeight(i) {
  const v = state.spectrum[i] || 0
  return Math.max(4, v * 72) + 'px'
}

// ----- Progress seeking (click + drag) -----
const progressTrack = ref(null)
let draggingProgress = false

function seekFromEvent(e) {
  const el = progressTrack.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  seekFraction((e.clientX - rect.left) / rect.width)
}
function onProgressDown(e) {
  draggingProgress = true
  seekFromEvent(e)
  window.addEventListener('pointermove', onProgressMove)
  window.addEventListener('pointerup', onProgressUp)
}
function onProgressMove(e) {
  if (draggingProgress) seekFromEvent(e)
}
function onProgressUp() {
  draggingProgress = false
  window.removeEventListener('pointermove', onProgressMove)
  window.removeEventListener('pointerup', onProgressUp)
}

// ----- Volume seeking (click + drag) -----
const volumeSlider = ref(null)
let draggingVolume = false

function volFromEvent(e) {
  const el = volumeSlider.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  setVolume((e.clientX - rect.left) / rect.width)
}
function onVolDown(e) {
  draggingVolume = true
  volFromEvent(e)
  window.addEventListener('pointermove', onVolMove)
  window.addEventListener('pointerup', onVolUp)
}
function onVolMove(e) {
  if (draggingVolume) volFromEvent(e)
}
function onVolUp() {
  draggingVolume = false
  window.removeEventListener('pointermove', onVolMove)
  window.removeEventListener('pointerup', onVolUp)
}

const repeatTitle = computed(() =>
  state.repeat === 'one' ? t('player.repeatOne') : state.repeat === 'all' ? t('player.repeatAll') : t('player.repeat')
)
</script>

<template>
  <main class="main-content">
    <ThemeSwitcher />

    <div class="album-section">
      <div class="vinyl-container">
        <div class="vinyl-disc" :class="{ spinning: state.isPlaying }">
          <div class="vinyl-grooves"></div>
          <div class="vinyl-label">
            <img v-if="currentTrack?.coverUrl" class="vinyl-cover" :src="currentTrack.coverUrl" alt="" />
            <div class="vinyl-hole"></div>
          </div>
        </div>
      </div>

      <div class="now-playing-info">
        <div class="now-playing-title">{{ title }}</div>
        <div class="now-playing-artist">{{ artist }}</div>
      </div>
    </div>

    <div class="spectrum-container">
      <div
        v-for="i in bars"
        :key="i"
        class="spectrum-bar"
        :style="{ height: barHeight(i) }"
      ></div>
    </div>

    <div class="progress-section">
      <span class="time-label">{{ formatTime(state.currentTime) }}</span>
      <div ref="progressTrack" class="progress-track" @pointerdown="onProgressDown">
        <div class="progress-fill" :style="{ width: progressPct + '%' }">
          <div class="progress-knob"></div>
        </div>
      </div>
      <span class="time-label right">{{ formatTime(state.duration) }}</span>
    </div>

    <div class="controls-section">
      <button
        class="ctrl-btn small"
        :class="{ 'is-on': state.shuffle }"
        :title="state.shuffle ? t('player.shuffleOn') : t('player.shuffleOff')"
        :aria-label="state.shuffle ? t('player.shuffleOn') : t('player.shuffleOff')"
        @click="toggleShuffle"
      >
        <!-- 随机播放：交叉箭头 -->
        <svg v-if="state.shuffle" class="icon" viewBox="0 0 24 24">
          <path d="M16 3h5v5M4 20L21 3M21 16v5h-5M15 15l6 6M4 4l5 5" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" fill="none" />
        </svg>
        <!-- 顺序播放：平行直箭头 -->
        <svg v-else class="icon" viewBox="0 0 24 24">
          <path d="M3 7h13M3 17h13M14 4l3 3-3 3M14 14l3 3-3 3" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" fill="none" />
        </svg>
      </button>
      <button class="ctrl-btn medium" :title="t('player.prev')" :aria-label="t('player.prev')" @click="prev">
        <svg class="icon large" viewBox="0 0 24 24">
          <path d="M19 20L9 12l10-8v16zM5 19V5" stroke="currentColor" stroke-width="2" stroke-linecap="round"
            stroke-linejoin="round" fill="none" />
        </svg>
      </button>
      <button class="ctrl-btn play-btn" :title="t('player.playPause')" :aria-label="t('player.playPauseAria')" @click="togglePlay">
        <svg v-if="!state.isPlaying" class="icon large" viewBox="0 0 24 24">
          <path d="M6 4l14 8-14 8V4z" fill="currentColor" />
        </svg>
        <svg v-else class="icon large" viewBox="0 0 24 24">
          <rect x="5" y="4" width="4" height="16" rx="1" fill="currentColor" />
          <rect x="15" y="4" width="4" height="16" rx="1" fill="currentColor" />
        </svg>
      </button>
      <button class="ctrl-btn medium" :title="t('player.next')" :aria-label="t('player.next')" @click="next(false)">
        <svg class="icon large" viewBox="0 0 24 24">
          <path d="M5 4l10 8-10 8V4zM19 5v14" stroke="currentColor" stroke-width="2" stroke-linecap="round"
            stroke-linejoin="round" fill="none" />
        </svg>
      </button>
      <button
        class="ctrl-btn small"
        :class="{ 'is-on': state.repeat !== 'off' }"
        :title="repeatTitle"
        :aria-label="t('player.repeat')"
        @click="cycleRepeat"
      >
        <svg class="icon" viewBox="0 0 24 24">
          <path d="M17 1l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round"
            stroke-linejoin="round" fill="none" />
          <path d="M3 11V9a4 4 0 0 1 4-4h14M7 23l-4-4 4-4" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" fill="none" />
          <path d="M21 13v2a4 4 0 0 1-4 4H3" stroke="currentColor" stroke-width="2" stroke-linecap="round"
            stroke-linejoin="round" fill="none" />
        </svg>
        <span v-if="state.repeat === 'one'" class="repeat-one">1</span>
      </button>
    </div>

    <div class="volume-section">
      <button class="ctrl-btn small" :aria-label="t('player.volume')" @click="toggleMute">
        <svg v-if="!state.muted && state.volume > 0" class="icon" viewBox="0 0 24 24">
          <path d="M11 5L6 9H2v6h4l5 4V5z" fill="currentColor" />
          <path d="M15.54 8.46a5 5 0 0 1 0 7.07M19.07 4.93a10 10 0 0 1 0 14.14" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" fill="none" />
        </svg>
        <svg v-else class="icon" viewBox="0 0 24 24">
          <path d="M11 5L6 9H2v6h4l5 4V5z" fill="currentColor" />
          <path d="M17 9l6 6M23 9l-6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" fill="none" />
        </svg>
      </button>
      <div ref="volumeSlider" class="volume-slider" @pointerdown="onVolDown">
        <div class="volume-fill" :style="{ width: volumePct + '%' }"></div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 26px;
  padding: 32px 40px;
  position: relative;
  overflow: hidden;
}

.main-content::before {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-radial-gradient(circle at 50% 40%, transparent 0px, transparent 38px, var(--border) 39px, transparent 40px);
  opacity: 0.4;
  pointer-events: none;
}

.album-section {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.vinyl-container {
  position: relative;
  width: 240px;
  height: 240px;
}

.vinyl-disc {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: conic-gradient(from 0deg,
      color-mix(in srgb, var(--seed-surface) 90%, #000 10%),
      color-mix(in srgb, var(--seed-surface) 70%, var(--seed-fg) 5%),
      color-mix(in srgb, var(--seed-surface) 90%, #000 10%),
      color-mix(in srgb, var(--seed-surface) 70%, var(--seed-fg) 5%),
      color-mix(in srgb, var(--seed-surface) 90%, #000 10%));
  box-shadow:
    0 0 0 3px color-mix(in srgb, var(--seed-fg) 10%, transparent),
    0 8px 32px color-mix(in srgb, var(--shadow-color) 60%, transparent),
    inset 0 0 60px color-mix(in srgb, #000 30%, transparent);
  animation: spin 8s linear infinite;
  animation-play-state: paused;
}

.vinyl-disc.spinning {
  animation-play-state: running;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.vinyl-grooves {
  position: absolute;
  inset: 20px;
  border-radius: 50%;
  border: 1px solid color-mix(in srgb, var(--seed-fg) 6%, transparent);
  box-shadow:
    inset 0 0 0 10px color-mix(in srgb, var(--seed-fg) 3%, transparent),
    inset 0 0 0 20px color-mix(in srgb, var(--seed-fg) 2%, transparent),
    inset 0 0 0 30px color-mix(in srgb, var(--seed-fg) 3%, transparent),
    inset 0 0 0 40px color-mix(in srgb, var(--seed-fg) 2%, transparent);
}

.vinyl-label {
  position: absolute;
  inset: 66px;
  border-radius: 50%;
  background: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-shadow: inset 0 2px 8px color-mix(in srgb, #000 25%, transparent);
}

.vinyl-cover {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.vinyl-hole {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--bg);
  box-shadow: inset 0 1px 3px color-mix(in srgb, #000 40%, transparent);
  z-index: 1;
}

.spectrum-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 480px;
  height: 76px;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 3px;
  padding: 0 20px;
}

.spectrum-bar {
  width: 4px;
  border-radius: 2px 2px 0 0;
  background: linear-gradient(to top, var(--spectrum-bar), var(--spectrum-bar-alt));
  transition: height 0.09s ease;
  opacity: 0.85;
}

.spectrum-bar:nth-child(even) {
  background: linear-gradient(to top, var(--spectrum-bar-alt), var(--spectrum-bar));
}

.now-playing-info {
  position: relative;
  z-index: 1;
  text-align: center;
}

.now-playing-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--fg);
  margin-bottom: 6px;
  letter-spacing: -0.2px;
  max-width: 420px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.now-playing-artist {
  font-size: 14px;
  color: var(--text-secondary);
}

.progress-section {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 480px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.time-label {
  font-size: 11px;
  color: var(--text-tertiary);
  min-width: 36px;
  font-variant-numeric: tabular-nums;
}

.time-label.right {
  text-align: right;
}

.progress-track {
  flex: 1;
  height: 5px;
  background: var(--progress-track);
  border-radius: 3px;
  position: relative;
  cursor: pointer;
  overflow: visible;
  touch-action: none;
}

.progress-fill {
  height: 100%;
  background: var(--progress-fill);
  border-radius: 3px;
  position: relative;
}

.progress-knob {
  position: absolute;
  right: -7px;
  top: 50%;
  transform: translateY(-50%);
  width: var(--knob-size);
  height: var(--knob-size);
  border-radius: 50%;
  background: var(--fg);
  box-shadow: 0 1px 4px color-mix(in srgb, #000 30%, transparent);
  opacity: 0;
  transition: opacity 0.15s;
}

.progress-track:hover .progress-knob {
  opacity: 1;
}

.controls-section {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 20px;
}

.ctrl-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--fg);
  cursor: pointer;
  transition: transform 0.1s, color 0.15s, background 0.15s;
  border-radius: 50%;
  position: relative;
}

.ctrl-btn:hover {
  color: var(--primary);
  transform: scale(1.08);
}

.ctrl-btn:active {
  transform: scale(0.95);
}

.ctrl-btn.is-on {
  color: var(--primary);
}

.ctrl-btn.small {
  width: 36px;
  height: 36px;
}

.ctrl-btn.medium {
  width: 44px;
  height: 44px;
}

.ctrl-btn.play-btn {
  width: 60px;
  height: 60px;
  background: var(--primary);
  color: var(--bg);
  box-shadow: 0 4px 16px var(--primary-glow);
}

.ctrl-btn.play-btn:hover {
  background: color-mix(in srgb, var(--seed-primary) 85%, var(--seed-fg) 15%);
  color: var(--bg);
  transform: scale(1.06);
}

.repeat-one {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -40%);
  font-size: 9px;
  font-weight: 700;
}

.volume-section {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 10px;
}

.volume-slider {
  width: 100px;
  height: 4px;
  background: var(--progress-track);
  border-radius: 2px;
  position: relative;
  cursor: pointer;
  touch-action: none;
}

.volume-fill {
  height: 100%;
  background: var(--accent);
  border-radius: 2px;
}

.icon {
  width: 20px;
  height: 20px;
  fill: currentColor;
}

.icon.large {
  width: 26px;
  height: 26px;
}
</style>
