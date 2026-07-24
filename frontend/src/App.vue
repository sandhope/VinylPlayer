<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import TitleBar from './components/TitleBar.vue'
import Sidebar from './components/Sidebar.vue'
import PlayerMain from './components/PlayerMain.vue'
import StatusBar from './components/StatusBar.vue'
import EqualizerPanel from './components/EqualizerPanel.vue'
import LyricsPanel from './components/LyricsPanel.vue'
import { usePlayer, setBase } from './composables/usePlayer'
import { useTheme } from './composables/useTheme'
import { GetInitialTracks, OpenFolder, OpenFiles, MediaBaseURL, RemoveTrack, ClearLibrary } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

const { state, setTracks, addTracks, removeTrack, clearTracks, loadIndex, togglePlay, next, prev } = usePlayer()
const { init: initTheme } = useTheme()

const eqOpen = ref(false)
const lyricsOpen = ref(false)
const busy = ref(false)

function toggleEq() {
  eqOpen.value = !eqOpen.value
  if (eqOpen.value) lyricsOpen.value = false
}
function toggleLyrics() {
  lyricsOpen.value = !lyricsOpen.value
  if (lyricsOpen.value) eqOpen.value = false
}

async function onAddFolder() {
  if (busy.value) return
  busy.value = true
  try {
    const tracks = await OpenFolder()
    if (tracks && tracks.length) {
      const wasEmpty = state.tracks.length === 0
      addTracks(tracks)
      if (wasEmpty) loadIndex(0, false)
    }
  } catch (e) {
    console.warn('OpenFolder failed:', e)
  } finally {
    busy.value = false
  }
}

async function onAddFiles() {
  if (busy.value) return
  busy.value = true
  try {
    const tracks = await OpenFiles()
    if (tracks && tracks.length) {
      const wasEmpty = state.tracks.length === 0
      addTracks(tracks)
      if (wasEmpty) loadIndex(0, false)
    }
  } catch (e) {
    console.warn('OpenFiles failed:', e)
  } finally {
    busy.value = false
  }
}

function onRemoveTrack(id) {
  removeTrack(id)
  RemoveTrack(id).catch((e) => console.warn('RemoveTrack failed:', e))
}

function onClearAll() {
  clearTracks()
  ClearLibrary().catch((e) => console.warn('ClearLibrary failed:', e))
}

function onKeydown(e) {
  const tag = e.target && e.target.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA') return
  switch (e.code) {
    case 'Space':
      e.preventDefault()
      togglePlay()
      break
    case 'ArrowRight':
      if (e.ctrlKey || e.metaKey) next(false)
      break
    case 'ArrowLeft':
      if (e.ctrlKey || e.metaKey) prev()
      break
  }
}

let unbindDrop = null

onMounted(async () => {
  initTheme()
  window.addEventListener('keydown', onKeydown)
  // Native OS file drops (audio files or folders) are imported by the backend,
  // which emits the freshly scanned tracks back to us.
  unbindDrop = EventsOn('tracks:dropped', (tracks) => {
    if (tracks && tracks.length) {
      const wasEmpty = state.tracks.length === 0
      addTracks(tracks)
      if (wasEmpty) loadIndex(0, false)
    }
  })
  try {
    // Resolve the media server base URL before ingesting tracks so their
    // audio/cover/lyric URLs are built against the correct origin.
    setBase(await MediaBaseURL())
    const tracks = await GetInitialTracks()
    if (tracks && tracks.length) {
      setTracks(tracks)
      loadIndex(0, false) // preload first track, wait for user gesture to play
    }
  } catch (e) {
    console.warn('initial load failed:', e)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  if (unbindDrop) unbindDrop()
})
</script>

<template>
  <TitleBar />
  <div class="app-body">
    <Sidebar
      :busy="busy"
      @add-folder="onAddFolder"
      @add-files="onAddFiles"
      @remove-track="onRemoveTrack"
      @clear-all="onClearAll"
    />
    <div class="stage">
      <PlayerMain />
      <Transition name="panel-slide">
        <EqualizerPanel v-if="eqOpen" @close="eqOpen = false" />
      </Transition>
      <Transition name="panel-slide">
        <LyricsPanel v-if="lyricsOpen" @close="lyricsOpen = false" />
      </Transition>
    </div>
  </div>
  <StatusBar
    :eq-open="eqOpen"
    :lyrics-open="lyricsOpen"
    @toggle-eq="toggleEq"
    @toggle-lyrics="toggleLyrics"
  />
</template>

<style scoped>
.app-body {
  display: flex;
  flex: 1;
  overflow: hidden;
  /* Opt the whole content area in as a native file-drop target (Wails). */
  --wails-drop-target: drop;
}

.stage {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
}

.panel-slide-enter-active,
.panel-slide-leave-active {
  transition: transform 0.28s ease, opacity 0.28s ease;
}

.panel-slide-enter-from,
.panel-slide-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>
