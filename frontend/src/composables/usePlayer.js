import { reactive, computed } from 'vue'

// ---- Equalizer configuration ----
// Six peaking bands spanning the audible range. Gains are in dB (-12..+12).
export const EQ_FREQUENCIES = [60, 170, 350, 1000, 3500, 10000]

export const EQ_PRESETS = {
  flat:      { label: '平坦',   gains: [0, 0, 0, 0, 0, 0] },
  pop:       { label: '流行',   gains: [-1, 2, 4, 3, 1, -1] },
  rock:      { label: '摇滚',   gains: [4, 2, -1, -1, 2, 4] },
  jazz:      { label: '爵士',   gains: [3, 1, 0, 1, 2, 3] },
  classical: { label: '古典',   gains: [4, 2, 0, 0, 1, 3] },
  bass:      { label: '重低音', gains: [7, 5, 2, 0, 0, 0] },
  vocal:     { label: '人声',   gains: [-2, -1, 2, 4, 3, 0] },
}

const BAR_COUNT = 48

// Module-level singleton state so every component shares one player instance.
const state = reactive({
  tracks: [],
  currentIndex: -1,
  isPlaying: false,
  currentTime: 0,
  duration: 0,
  volume: 0.7,
  muted: false,
  shuffle: false,
  repeat: 'off', // 'off' | 'all' | 'one'
  spectrum: new Array(BAR_COUNT).fill(0),
  eqEnabled: false,
  eqGains: [...EQ_PRESETS.flat.gains],
  eqPreset: 'flat',
  lyrics: [],       // [{ time, text }]
  lyricIndex: -1,
  loading: false,
})

// Base URL of the loopback media server (e.g. "http://127.0.0.1:52341").
// Set once at startup via setBase() before any tracks are ingested.
let BASE = ''

// setBase records the media server base URL. Call before loading tracks.
export function setBase(url) {
  BASE = url || ''
}

// absoluteUrl turns a backend-relative media path into an absolute URL against
// the media server, leaving already-absolute URLs untouched.
function absoluteUrl(u) {
  if (!u) return ''
  return /^https?:\/\//i.test(u) ? u : BASE + u
}

// ingest normalises a raw track from Go: media/cover/lyric URLs are made
// absolute so they resolve to the media server regardless of the page origin.
function ingest(t) {
  return {
    ...t,
    url: absoluteUrl(t.url),
    coverUrl: absoluteUrl(t.coverUrl),
    lyricUrl: absoluteUrl(t.lyricUrl),
    duration: 0,
  }
}

// ---- Playback progress (resume where the user left off) ----
// A map of trackId -> last playback position (seconds), seeded from the backend
// at startup and kept in sync as the user listens. Persistence is delegated to
// an injected backend so this module stays free of Wails imports.
const progress = new Map()
let progressBackend = { save: () => {}, clear: () => {} }
let lastSaveAt = 0
let pendingSeek = 0

// Only remember/restore positions meaningfully into a track and not basically
// finished.
const RESUME_MIN = 5
const RESUME_END_GUARD = 5

// hydrateProgress seeds the in-memory positions from a { id: seconds } map.
export function hydrateProgress(map) {
  progress.clear()
  if (!map) return
  for (const [id, sec] of Object.entries(map)) {
    const s = Number(sec)
    if (id && s > RESUME_MIN) progress.set(id, s)
  }
}

// setProgressBackend injects the persistence callbacks (backed by Wails).
export function setProgressBackend(backend) {
  if (backend) progressBackend = backend
}

// rememberProgress stores the current position in memory and, throttled (or
// when forced), pushes it to the backend.
function rememberProgress(force = false) {
  const track = state.tracks[state.currentIndex]
  if (!track || !audio) return
  const t = audio.currentTime
  const d = state.duration
  if (t < RESUME_MIN || (d && t > d - RESUME_END_GUARD)) return
  progress.set(track.id, t)
  const now = Date.now()
  if (force || now - lastSaveAt > 4000) {
    lastSaveAt = now
    progressBackend.save(track.id, t)
  }
}

// clearProgress forgets a track's saved position (used when it finishes).
function clearProgress(track) {
  if (!track) return
  progress.delete(track.id)
  progressBackend.clear(track.id)
}

// flushProgress force-saves the current position (e.g. on app exit).
export function flushProgress() {
  rememberProgress(true)
}

// Web Audio objects (created lazily on first user-initiated playback).
let audio = null
let audioCtx = null
let sourceNode = null
let analyser = null
let eqNodes = []
let freqData = null
let rafId = null

function ensureAudio() {
  if (audio) return
  audio = new Audio()
  audio.preload = 'auto'
  audio.crossOrigin = 'anonymous'
  audio.volume = state.volume

  audio.addEventListener('loadedmetadata', () => {
    state.duration = audio.duration || 0
    const t = state.tracks[state.currentIndex]
    if (t) t.duration = state.duration
    // Resume from the saved position if we have one for this track.
    if (pendingSeek > 0 && state.duration && pendingSeek < state.duration - RESUME_END_GUARD) {
      try {
        audio.currentTime = pendingSeek
      } catch (e) {
        /* seeking not ready yet; ignore */
      }
      state.currentTime = pendingSeek
    }
    pendingSeek = 0
  })
  audio.addEventListener('timeupdate', () => {
    state.currentTime = audio.currentTime
    updateLyricIndex()
    rememberProgress()
  })
  audio.addEventListener('play', () => {
    state.isPlaying = true
    startSpectrum()
  })
  audio.addEventListener('pause', () => {
    state.isPlaying = false
    stopSpectrum()
    rememberProgress(true)
  })
  audio.addEventListener('ended', onEnded)
}

// Build the Web Audio graph: source -> EQ chain -> analyser -> destination.
function ensureGraph() {
  if (audioCtx) return
  try {
    audioCtx = new (window.AudioContext || window.webkitAudioContext)()
    sourceNode = audioCtx.createMediaElementSource(audio)

    eqNodes = EQ_FREQUENCIES.map((freq, i) => {
      const f = audioCtx.createBiquadFilter()
      f.type = 'peaking'
      f.frequency.value = freq
      f.Q.value = 1.1
      f.gain.value = state.eqEnabled ? state.eqGains[i] : 0
      return f
    })

    analyser = audioCtx.createAnalyser()
    analyser.fftSize = 256
    analyser.smoothingTimeConstant = 0.8
    freqData = new Uint8Array(analyser.frequencyBinCount)

    // Wire the chain.
    let node = sourceNode
    for (const eq of eqNodes) {
      node.connect(eq)
      node = eq
    }
    node.connect(analyser)
    analyser.connect(audioCtx.destination)
  } catch (e) {
    console.warn('Web Audio unavailable, spectrum/EQ disabled:', e)
  }
}

function startSpectrum() {
  if (!analyser || rafId) return
  const render = () => {
    rafId = requestAnimationFrame(render)
    if (!analyser) return
    analyser.getByteFrequencyData(freqData)
    const bins = freqData.length
    const out = state.spectrum
    for (let i = 0; i < BAR_COUNT; i++) {
      // Logarithmic-ish mapping favouring lower frequencies for a musical look.
      const idx = Math.floor(Math.pow(i / BAR_COUNT, 1.4) * bins)
      const v = freqData[Math.min(idx, bins - 1)] / 255
      out[i] = v
    }
  }
  render()
}

function stopSpectrum() {
  if (rafId) {
    cancelAnimationFrame(rafId)
    rafId = null
  }
  state.spectrum = new Array(BAR_COUNT).fill(0)
}

function mediaUrl(track) {
  return track.url
}

async function loadIndex(index, autoplay = true) {
  if (index < 0 || index >= state.tracks.length) return
  ensureAudio()
  // Save the outgoing track's position before switching away.
  if (state.currentIndex >= 0 && state.currentIndex !== index) {
    rememberProgress(true)
  }
  state.currentIndex = index
  const track = state.tracks[index]
  audio.src = mediaUrl(track)
  state.currentTime = 0
  state.duration = track.duration || 0
  // Queue a resume seek; applied once loadedmetadata reports the duration.
  const saved = progress.get(track.id) || 0
  pendingSeek = saved > RESUME_MIN ? saved : 0
  loadLyrics(track)
  if (autoplay) {
    await play()
  }
}

async function play() {
  ensureAudio()
  if (state.currentIndex < 0 && state.tracks.length) {
    await loadIndex(0, false)
  }
  ensureGraph()
  if (audioCtx && audioCtx.state === 'suspended') {
    await audioCtx.resume()
  }
  try {
    await audio.play()
  } catch (e) {
    console.warn('playback failed:', e)
  }
}

function pause() {
  if (audio) audio.pause()
}

function togglePlay() {
  if (state.currentIndex < 0 && state.tracks.length) {
    loadIndex(0, true)
    return
  }
  if (state.isPlaying) pause()
  else play()
}

function onEnded() {
  clearProgress(state.tracks[state.currentIndex])
  if (state.repeat === 'one') {
    audio.currentTime = 0
    play()
    return
  }
  next(true)
}

function pickNextIndex(dir) {
  const n = state.tracks.length
  if (n === 0) return -1
  if (state.shuffle) {
    if (n === 1) return 0
    let r
    do {
      r = Math.floor(Math.random() * n)
    } while (r === state.currentIndex)
    return r
  }
  return (state.currentIndex + dir + n) % n
}

function next(auto = false) {
  const n = state.tracks.length
  if (n === 0) return
  if (!state.shuffle && auto && state.repeat === 'off' && state.currentIndex === n - 1) {
    // Reached the end of the list with no repeat: stop.
    pause()
    if (audio) audio.currentTime = 0
    return
  }
  loadIndex(pickNextIndex(1), true)
}

function prev() {
  if (state.tracks.length === 0) return
  // Restart current track if more than 3s in, else go to previous.
  if (state.currentTime > 3 && audio) {
    audio.currentTime = 0
    return
  }
  loadIndex(pickNextIndex(-1), true)
}

function seekFraction(frac) {
  if (!audio || !state.duration) return
  const clamped = Math.max(0, Math.min(1, frac))
  audio.currentTime = clamped * state.duration
  state.currentTime = audio.currentTime
}

function setVolume(v) {
  state.volume = Math.max(0, Math.min(1, v))
  state.muted = state.volume === 0
  ensureAudio()
  audio.volume = state.volume
  audio.muted = false
}

function toggleMute() {
  ensureAudio()
  state.muted = !state.muted
  audio.muted = state.muted
}

function toggleShuffle() {
  state.shuffle = !state.shuffle
}

function cycleRepeat() {
  state.repeat = state.repeat === 'off' ? 'all' : state.repeat === 'all' ? 'one' : 'off'
}

// ---- Equalizer ----
function applyEqToNodes() {
  if (!eqNodes.length) return
  eqNodes.forEach((node, i) => {
    node.gain.value = state.eqEnabled ? state.eqGains[i] : 0
  })
}

function setEqEnabled(on) {
  state.eqEnabled = on
  applyEqToNodes()
}

function setEqGain(index, value) {
  state.eqGains[index] = value
  state.eqPreset = 'custom'
  applyEqToNodes()
}

function applyEqPreset(name) {
  const preset = EQ_PRESETS[name]
  if (!preset) return
  state.eqPreset = name
  state.eqGains = [...preset.gains]
  if (!state.eqEnabled) state.eqEnabled = true
  applyEqToNodes()
}

// ---- Lyrics (.lrc) ----
async function loadLyrics(track) {
  state.lyrics = []
  state.lyricIndex = -1
  if (!track.lyricUrl) return
  try {
    const res = await fetch(track.lyricUrl)
    if (!res.ok) return
    const text = await res.text()
    state.lyrics = parseLrc(text)
  } catch (e) {
    /* no lyrics */
  }
}

function parseLrc(text) {
  const lines = []
  const re = /\[(\d{1,2}):(\d{1,2})(?:[.:](\d{1,3}))?\]/g
  for (const raw of text.split(/\r?\n/)) {
    const content = raw.replace(re, '').trim()
    let m
    re.lastIndex = 0
    while ((m = re.exec(raw)) !== null) {
      const min = parseInt(m[1], 10)
      const sec = parseInt(m[2], 10)
      const ms = m[3] ? parseInt(m[3].padEnd(3, '0'), 10) : 0
      const time = min * 60 + sec + ms / 1000
      if (content) lines.push({ time, text: content })
    }
  }
  lines.sort((a, b) => a.time - b.time)
  return lines
}

function updateLyricIndex() {
  if (!state.lyrics.length) return
  const t = state.currentTime
  let idx = -1
  for (let i = 0; i < state.lyrics.length; i++) {
    if (state.lyrics[i].time <= t) idx = i
    else break
  }
  if (idx !== state.lyricIndex) state.lyricIndex = idx
}

// ---- Library ----
function setTracks(tracks, { autoplayFirst = false } = {}) {
  state.tracks = tracks.map(ingest)
  state.currentIndex = -1
  if (autoplayFirst && tracks.length) {
    loadIndex(0, false)
  }
}

function addTracks(tracks) {
  const existing = new Set(state.tracks.map((t) => t.id))
  const fresh = tracks.filter((t) => !existing.has(t.id)).map(ingest)
  state.tracks.push(...fresh)
  return fresh.length
}

// resetPlayback tears down the current track (used when the playing track is
// removed or the library is cleared).
function resetPlayback() {
  pause()
  if (audio) audio.removeAttribute('src')
  state.currentIndex = -1
  state.currentTime = 0
  state.duration = 0
  state.isPlaying = false
  state.lyrics = []
  state.lyricIndex = -1
  stopSpectrum()
}

// removeTrack drops a single track and keeps playback coherent: removing the
// current track advances to whatever now occupies its slot; removing an earlier
// track shifts the current index so the same song keeps playing.
function removeTrack(id) {
  const idx = state.tracks.findIndex((t) => t.id === id)
  if (idx < 0) return
  const wasCurrent = idx === state.currentIndex
  const wasPlaying = state.isPlaying
  state.tracks.splice(idx, 1)

  if (state.tracks.length === 0) {
    resetPlayback()
    return
  }
  if (wasCurrent) {
    const newIdx = Math.min(idx, state.tracks.length - 1)
    state.currentIndex = -1 // force loadIndex to treat this as a fresh load
    loadIndex(newIdx, wasPlaying)
  } else if (idx < state.currentIndex) {
    state.currentIndex -= 1
  }
}

// clearTracks empties the whole playlist and stops playback.
function clearTracks() {
  state.tracks = []
  resetPlayback()
}

const currentTrack = computed(() =>
  state.currentIndex >= 0 ? state.tracks[state.currentIndex] : null
)

export function usePlayer() {
  return {
    state,
    currentTrack,
    setTracks,
    addTracks,
    removeTrack,
    clearTracks,
    loadIndex,
    play,
    pause,
    togglePlay,
    next,
    prev,
    seekFraction,
    setVolume,
    toggleMute,
    toggleShuffle,
    cycleRepeat,
    setEqEnabled,
    setEqGain,
    applyEqPreset,
    stopSpectrum,
  }
}
