<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { usePlayer } from '../composables/usePlayer'
import { useSettings } from '../composables/useSettings'
import { formatTime } from '../composables/format'

defineProps({
  busy: { type: Boolean, default: false },
})
const emit = defineEmits(['add-folder', 'add-files', 'remove-track', 'clear-all'])

const { state, loadIndex } = usePlayer()
const { t } = useSettings()

function thumbShade(i) {
  return 0.55 + (i % 4) * 0.12
}

// ---- List / Album / Artist views ----
// viewMode is remembered across restarts. 'list' is the flat playlist; 'album'
// and 'artist' group tracks under collapsible headers.
const viewMode = ref(localStorage.getItem('vp-view-mode') || 'list')
watch(viewMode, (v) => localStorage.setItem('vp-view-mode', v))

// Collapsed group keys, kept as a plain object so Vue tracks reassignments.
const collapsed = ref({})
function isCollapsed(key) {
  return !!collapsed.value[key]
}
function toggleGroup(key) {
  collapsed.value = { ...collapsed.value, [key]: !collapsed.value[key] }
}

// groups buckets the current tracks by album or artist, keeping each group's
// first available cover art for the header thumbnail.
const groups = computed(() => {
  const mode = viewMode.value
  const keyOf = (track) =>
    mode === 'album' ? track.album || t('sidebar.unknownAlbum') : track.artist || t('sidebar.unknownArtist')
  const map = new Map()
  for (const track of state.tracks) {
    const k = keyOf(track)
    let g = map.get(k)
    if (!g) {
      g = { key: k, name: k, cover: '', tracks: [] }
      map.set(k, g)
    }
    g.tracks.push(track)
    if (!g.cover && track.coverUrl) g.cover = track.coverUrl
  }
  return Array.from(map.values()).sort((a, b) => a.name.localeCompare(b.name, 'zh'))
})

// rows flattens the sidebar into one render list: plain tracks (list mode) or
// group headers interleaved with their tracks. Each track row carries its real
// index into state.tracks so playback maps correctly regardless of view.
const rows = computed(() => {
  if (viewMode.value === 'list') {
    return state.tracks.map((t, i) => ({ type: 'track', track: t, index: i }))
  }
  const indexById = new Map(state.tracks.map((t, i) => [t.id, i]))
  const out = []
  for (const g of groups.value) {
    out.push({ type: 'header', group: g })
    if (!isCollapsed(g.key)) {
      for (const t of g.tracks) {
        out.push({ type: 'track', track: t, index: indexById.get(t.id) })
      }
    }
  }
  return out
})

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
      <div class="header-left">
        <span class="sidebar-title">{{ t('sidebar.title') }}</span>
        <span class="track-count">{{ t('sidebar.trackCount', { n: state.tracks.length }) }}</span>
      </div>
      <div class="header-right">
        <button
          v-if="state.tracks.length"
          class="clear-btn"
          :class="{ confirming: confirmingClear }"
          :title="confirmingClear ? t('sidebar.clearConfirm') : t('sidebar.clear')"
          @click="onClearClick"
        >
          <template v-if="confirmingClear">{{ t('sidebar.confirmShort') }}</template>
          <svg v-else viewBox="0 0 24 24" width="14" height="14">
            <path d="M4 7h16M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2m1 0v12a1 1 0 0 1-1 1H8a1 1 0 0 1-1-1V7"
              stroke="currentColor" stroke-width="1.8" fill="none" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
        <div v-if="state.tracks.length" class="view-switch">
          <button
            :class="{ active: viewMode === 'list' }"
            :title="t('sidebar.viewList')"
            :aria-label="t('sidebar.viewList')"
            @click="viewMode = 'list'"
          >
            <svg viewBox="0 0 24 24" width="15" height="15">
              <path d="M8 6h12M8 12h12M8 18h12M4 6h.01M4 12h.01M4 18h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
            </svg>
          </button>
          <button
            :class="{ active: viewMode === 'album' }"
            :title="t('sidebar.viewAlbum')"
            :aria-label="t('sidebar.viewAlbum')"
            @click="viewMode = 'album'"
          >
            <svg viewBox="0 0 24 24" width="15" height="15">
              <circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.8" fill="none" />
              <circle cx="12" cy="12" r="2" fill="currentColor" />
            </svg>
          </button>
          <button
            :class="{ active: viewMode === 'artist' }"
            :title="t('sidebar.viewArtist')"
            :aria-label="t('sidebar.viewArtist')"
            @click="viewMode = 'artist'"
          >
            <svg viewBox="0 0 24 24" width="15" height="15">
              <circle cx="12" cy="8" r="3.2" stroke="currentColor" stroke-width="1.8" fill="none" />
              <path d="M5.5 19a6.5 6.5 0 0 1 13 0" stroke="currentColor" stroke-width="1.8" fill="none" stroke-linecap="round" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div class="sidebar-actions">
      <button class="add-btn" :disabled="busy" @click="emit('add-folder')">
        <svg viewBox="0 0 24 24" width="15" height="15">
          <path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"
            stroke="currentColor" stroke-width="1.8" fill="none" stroke-linejoin="round" />
        </svg>
        {{ t('sidebar.addFolder') }}
      </button>
      <button class="add-btn" :disabled="busy" @click="emit('add-files')">
        <svg viewBox="0 0 24 24" width="15" height="15">
          <path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8l-5-5z"
            stroke="currentColor" stroke-width="1.8" fill="none" stroke-linejoin="round" />
          <path d="M14 3v5h5" stroke="currentColor" stroke-width="1.8" fill="none" stroke-linejoin="round" />
        </svg>
        {{ t('sidebar.addFiles') }}
      </button>
    </div>

    <div class="playlist">
      <template v-for="row in rows" :key="row.type === 'header' ? 'h:' + row.group.key : row.track.id">
        <div
          v-if="row.type === 'header'"
          class="group-header"
          @click="toggleGroup(row.group.key)"
        >
          <span class="group-caret" :class="{ collapsed: isCollapsed(row.group.key) }">
            <svg viewBox="0 0 24 24" width="12" height="12">
              <path d="M9 6l6 6-6 6" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </span>
          <div class="group-thumb">
            <img v-if="row.group.cover" :src="row.group.cover" alt="" />
            <svg v-else viewBox="0 0 40 40">
              <rect width="40" height="40" fill="var(--seed-primary)" opacity="0.6" />
              <circle cx="20" cy="20" r="12" fill="var(--seed-accent)" opacity="0.5" />
              <circle cx="20" cy="20" r="4" fill="var(--seed-surface)" />
            </svg>
          </div>
          <div class="group-info">
            <div class="group-name">{{ row.group.name }}</div>
            <div class="group-meta">{{ t('sidebar.trackCount', { n: row.group.tracks.length }) }}</div>
          </div>
        </div>
        <div
          v-else
          class="track-item"
          :class="{ active: row.index === state.currentIndex, 'in-group': viewMode !== 'list' }"
          @dblclick="loadIndex(row.index, true)"
          @click="loadIndex(row.index, true)"
        >
          <span class="track-index">
            <template v-if="row.index === state.currentIndex && state.isPlaying">♪</template>
            <template v-else-if="viewMode === 'list'">{{ row.index + 1 }}</template>
          </span>
          <div class="track-thumb">
            <img v-if="row.track.coverUrl" :src="row.track.coverUrl" alt="" />
            <svg v-else viewBox="0 0 40 40">
              <rect width="40" height="40" fill="var(--seed-primary)" :opacity="thumbShade(row.index)" />
              <circle cx="20" cy="20" r="12" fill="var(--seed-accent)" opacity="0.5" />
              <circle cx="20" cy="20" r="4" fill="var(--seed-surface)" />
            </svg>
          </div>
          <div class="track-info">
            <div class="track-name">{{ row.track.title }}</div>
            <div class="track-artist">{{ row.track.artist }}</div>
          </div>
          <span class="track-duration">{{ row.track.duration ? formatTime(row.track.duration) : row.track.format }}</span>
          <button
            class="track-remove"
            :title="t('sidebar.removeTrack')"
            @click.stop="emit('remove-track', row.track.id)"
          >
            <svg viewBox="0 0 24 24" width="14" height="14">
              <path d="M6 6l12 12M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
            </svg>
          </button>
        </div>
      </template>

      <div v-if="!state.tracks.length" class="empty-hint">
        <p>{{ t('sidebar.emptyTitle') }}</p>
        <span>{{ t('sidebar.emptyHint') }}</span>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  --sidebar-width: 300px;
  width: var(--sidebar-width);
  background: var(--surface);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;
  transition: width 0.28s ease, background 0.4s ease;
}

/* Keep inner content laid out at full width while the panel collapses, so it
   gets clipped (wiped) instead of squished/reflowed. */
.sidebar > * {
  min-width: var(--sidebar-width);
  flex-shrink: 0;
}

.sidebar.collapsed {
  width: 0;
  min-width: 0;
  border-right: none;
  padding: 0;
}

.sidebar-header {
  padding: 16px 16px 12px;
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

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
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

.view-switch {
  display: flex;
  gap: 1px;
  padding: 2px;
  background: color-mix(in srgb, var(--seed-fg) 6%, transparent);
  border-radius: calc(var(--radius) * 0.55);
  flex-shrink: 0;
}

.view-switch button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 22px;
  color: var(--text-tertiary);
  background: transparent;
  border: none;
  border-radius: calc(var(--radius) * 0.4);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.view-switch button:hover {
  color: var(--fg);
}

.view-switch button.active {
  color: var(--fg);
  background: var(--surface);
  box-shadow: 0 1px 3px color-mix(in srgb, var(--shadow-color) 30%, transparent);
}

.group-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  margin-top: 2px;
  border-radius: calc(var(--radius) * 0.6);
  cursor: pointer;
  transition: background 0.15s;
}

.group-header:hover {
  background: color-mix(in srgb, var(--seed-fg) 6%, transparent);
}

.group-caret {
  display: flex;
  flex-shrink: 0;
  color: var(--text-tertiary);
  transform: rotate(90deg);
  transition: transform 0.18s ease;
}

.group-caret.collapsed {
  transform: rotate(0deg);
}

.group-thumb {
  width: 34px;
  height: 34px;
  border-radius: calc(var(--radius) * 0.5);
  overflow: hidden;
  flex-shrink: 0;
  background: var(--surface-sunken);
}

.group-thumb svg,
.group-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.group-info {
  flex: 1;
  min-width: 0;
}

.group-name {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--fg);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-meta {
  font-size: 10.5px;
  color: var(--text-tertiary);
  margin-top: 1px;
}

.track-item.in-group {
  padding-left: 22px;
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
