import { ref } from 'vue'

export const THEMES = [
  { id: 'retro', label: '复古怀旧', cls: 'retro' },
  { id: 'dark', label: '深色沉浸', cls: 'dark' },
  { id: 'light', label: '浅色简洁', cls: 'light' },
  { id: 'minimal', label: '极简现代', cls: 'minimal' },
]

const STORAGE_KEY = 'vinyl-player-theme'
const current = ref('retro')

function apply(theme) {
  current.value = theme
  const root = document.documentElement
  if (theme === 'retro') {
    root.removeAttribute('data-theme')
  } else {
    root.setAttribute('data-theme', theme)
  }
  try {
    localStorage.setItem(STORAGE_KEY, theme)
  } catch (e) {
    /* ignore */
  }
}

function init() {
  let saved = 'retro'
  try {
    saved = localStorage.getItem(STORAGE_KEY) || 'retro'
  } catch (e) {
    /* ignore */
  }
  apply(saved)
}

export function useTheme() {
  return { current, themes: THEMES, apply, init }
}
