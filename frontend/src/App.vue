<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import TitleBar from './components/TitleBar.vue'
import Sidebar from './components/Sidebar.vue'
import PlayerMain from './components/PlayerMain.vue'
import StatusBar from './components/StatusBar.vue'
import EqualizerPanel from './components/EqualizerPanel.vue'
import LyricsPanel from './components/LyricsPanel.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import AboutPanel from './components/AboutPanel.vue'
import { usePlayer, setBase, hydrateProgress, setProgressBackend, setRememberEnabled, flushProgress } from './composables/usePlayer'
import { useTheme } from './composables/useTheme'
import { useSettings } from './composables/useSettings'
import { GetInitialTracks, OpenFolder, OpenFiles, MediaBaseURL, RemoveTrack, ClearLibrary, GetProgress, SaveProgress, ClearProgress, AddPaths } from '../wailsjs/go/main/App'
import { OnFileDrop, OnFileDropOff, WindowFullscreen, WindowUnfullscreen, WindowIsFullscreen, WindowSetAlwaysOnTop } from '../wailsjs/runtime/runtime'
import { Quit } from '../wailsjs/go/main/App'

const { state, setTracks, addTracks, removeTrack, clearTracks, loadIndex, togglePlay, next, prev } = usePlayer()
const { init: initTheme, current, themes, apply: applyTheme } = useTheme()
const { settings, t } = useSettings()

const eqOpen = ref(false)
const lyricsOpen = ref(false)
const settingsOpen = ref(false)
const aboutOpen = ref(false)
const busy = ref(false)
const SIDEBAR_KEY = 'vinyl-player-sidebar-collapsed'
const sidebarOpen = ref(localStorage.getItem(SIDEBAR_KEY) !== '1')

// Keep the player's position-memory feature in sync with the user's setting,
// applied immediately so it's in effect before the first track loads.
watch(() => settings.rememberProgress, (v) => setRememberEnabled(v), { immediate: true })

function toggleEq() {
  eqOpen.value = !eqOpen.value
  if (eqOpen.value) lyricsOpen.value = false
}
function toggleLyrics() {
  lyricsOpen.value = !lyricsOpen.value
  if (lyricsOpen.value) eqOpen.value = false
}
function openLyrics() {
  lyricsOpen.value = true
  eqOpen.value = false
}
function toggleSidebar() {
  sidebarOpen.value = !sidebarOpen.value
  try {
    localStorage.setItem(SIDEBAR_KEY, sidebarOpen.value ? '0' : '1')
  } catch (e) {
    /* ignore */
  }
}

// ---- Window-level right-click menu on the play area (mirrors ReelPlayer) ----
const mainMenu = ref({ open: false, x: 0, y: 0 })
const themeSubmenu = ref(false)
const alwaysOnTop = ref(false)

function openMainMenu(e) {
  e.preventDefault()
  mainMenu.value = { open: true, x: e.clientX, y: e.clientY }
}
function closeMainMenu() {
  mainMenu.value = { ...mainMenu.value, open: false }
}
const mainMenuStyle = computed(() => ({
  left: mainMenu.value.x + 'px',
  top: mainMenu.value.y + 'px',
}))
async function mainMenuFullscreen() {
  closeMainMenu()
  const isFull = await WindowIsFullscreen()
  if (isFull) WindowUnfullscreen()
  else WindowFullscreen()
}
function mainMenuToggleOnTop() {
  alwaysOnTop.value = !alwaysOnTop.value
  WindowSetAlwaysOnTop(alwaysOnTop.value)
  closeMainMenu()
}
function mainMenuTheme(id) {
  applyTheme(id)
  closeMainMenu()
}
function mainMenuQuit() {
  closeMainMenu()
  Quit()
}
function onMainMenuKey(e) {
  if (e.key === 'Escape') closeMainMenu()
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
      // Add any new tracks, then always switch to (and play) the first opened
      // track — even when it is already in the playlist, so re-opening an
      // existing track still jumps to it instead of doing nothing.
      addTracks(tracks)
      const idx = state.tracks.findIndex((t) => t.id === tracks[0].id)
      if (idx >= 0) loadIndex(idx, true)
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

// Registering OnFileDrop on the JS side is what installs Wails' window-level
// dragover/drop listeners (which preventDefault). Without it, WebView2 falls
// back to its native behavior and opens dropped files in a popup window.
async function onFilesDropped(x, y, paths) {
  if (!paths || !paths.length) return
  try {
    const tracks = await AddPaths(paths)
    if (tracks && tracks.length) {
      // Same semantics as onAddFiles: switch to (and play) the first dropped
      // track, even if it was already in the playlist.
      addTracks(tracks)
      const idx = state.tracks.findIndex((t) => t.id === tracks[0].id)
      if (idx >= 0) loadIndex(idx, true)
    }
  } catch (e) {
    console.warn('AddPaths failed:', e)
  }
}

onMounted(async () => {
  initTheme()
  window.addEventListener('keydown', onKeydown)
  // useDropTarget=false: accept drops anywhere in the window, not only over
  // elements carrying the --wails-drop-target CSS property.
  OnFileDrop(onFilesDropped, false)
  // Wire playback-progress persistence and seed saved positions before the
  // first track loads so it can resume where the user left off.
  setProgressBackend({
    save: (id, sec) => SaveProgress(id, sec).catch(() => {}),
    clear: (id) => ClearProgress(id).catch(() => {}),
  })
  try {
    // Resolve the media server base URL before ingesting tracks so their
    // audio/cover/lyric URLs are built against the correct origin.
    setBase(await MediaBaseURL())
    try {
      hydrateProgress(await GetProgress())
    } catch (e) {
      /* no saved progress */
    }
    const tracks = await GetInitialTracks()
    if (tracks && tracks.length) {
      setTracks(tracks)
      loadIndex(0, false) // preload first track, wait for user gesture to play
    }
  } catch (e) {
    console.warn('initial load failed:', e)
  }
  // Best-effort save when the window is closing (periodic + pause saves cover
  // the rest).
  window.addEventListener('beforeunload', flushProgress)
  document.addEventListener('click', closeMainMenu)
  window.addEventListener('keydown', onMainMenuKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('keydown', onMainMenuKey)
  window.removeEventListener('beforeunload', flushProgress)
  document.removeEventListener('click', closeMainMenu)
  flushProgress()
  OnFileDropOff()
})
</script>

<template>
  <TitleBar @open-settings="settingsOpen = true" @open-about="aboutOpen = true" />
  <div class="app-body">
    <Sidebar
      :busy="busy"
      :class="{ collapsed: !sidebarOpen }"
      @add-folder="onAddFolder"
      @add-files="onAddFiles"
      @remove-track="onRemoveTrack"
      @clear-all="onClearAll"
    />
    <div class="stage" @contextmenu.prevent="openMainMenu($event)">
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
      :sidebar-open="sidebarOpen"
      @toggle-eq="toggleEq"
      @toggle-lyrics="toggleLyrics"
      @toggle-sidebar="toggleSidebar"
      @open-lyrics="openLyrics"
    />
  <Transition name="fade">
    <SettingsPanel v-if="settingsOpen" @close="settingsOpen = false" />
  </Transition>
  <Transition name="fade">
    <AboutPanel v-if="aboutOpen" @close="aboutOpen = false" />
  </Transition>

  <Teleport to="body">
    <div v-if="mainMenu.open" class="ctx-overlay" @click="closeMainMenu" @contextmenu.prevent="closeMainMenu">
      <div class="ctx-menu" :style="mainMenuStyle" @click.stop>
        <button class="ctx-item" @click="onAddFiles(); closeMainMenu()">{{ t('ctx.openFiles') }}</button>
        <button class="ctx-item" @click="onAddFolder(); closeMainMenu()">{{ t('ctx.openFolder') }}</button>
        <div class="ctx-sep"></div>
        <button class="ctx-item" @click="mainMenuFullscreen">{{ t('ctx.fullscreen') }}</button>
        <button class="ctx-item" :class="{ active: alwaysOnTop }" @click="mainMenuToggleOnTop">
          {{ t('ctx.alwaysOnTop') }}<span v-if="alwaysOnTop" class="ctx-check">✓</span>
        </button>
        <div class="ctx-item ctx-sub" @mouseenter="themeSubmenu = true" @mouseleave="themeSubmenu = false">
          {{ t('ctx.theme') }}
          <svg class="ctx-arrow" viewBox="0 0 24 24" width="16" height="16">
            <path d="M9 6l6 6-6 6" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <div v-show="themeSubmenu" class="ctx-submenu">
            <button
              v-for="th in themes" :key="th.id"
              class="ctx-item" :class="{ active: current === th.id }"
              @click="mainMenuTheme(th.id)"
            >{{ th.label }}<span v-if="current === th.id" class="ctx-check">✓</span></button>
          </div>
        </div>
        <div class="ctx-sep"></div>
        <button class="ctx-item" @click="settingsOpen = true; closeMainMenu()">{{ t('ctx.settings') }}</button>
        <button class="ctx-item" @click="aboutOpen = true; closeMainMenu()">{{ t('ctx.about') }}</button>
        <div class="ctx-sep"></div>
        <button class="ctx-item" @click="mainMenuQuit">{{ t('ctx.quit') }}</button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.app-body {
  display: flex;
  flex: 1;
  overflow: hidden;
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

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
