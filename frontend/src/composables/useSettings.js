import { reactive, watch } from 'vue'
import { messages } from '../i18n/messages'

const LOCALE_KEY = 'vp-locale'
const REMEMBER_KEY = 'vp-remember-progress'

// Map a browser/OS locale tag (e.g. "zh-CN", "en-US", "ja") onto a locale we
// ship. Chinese systems get Chinese; everything else falls back to English.
function detectSystemLocale() {
  try {
    const tags =
      navigator.languages && navigator.languages.length
        ? navigator.languages
        : [navigator.language]
    for (const tag of tags) {
      if (!tag) continue
      if (tag.toLowerCase().startsWith('zh')) return 'zh'
    }
  } catch (e) {
    /* ignore */
  }
  return 'en'
}

// On first launch we detect the system language once and persist it; every
// later launch just reads the stored value, so the detection never runs again.
function initialLocale() {
  try {
    const saved = localStorage.getItem(LOCALE_KEY)
    if (saved && messages[saved]) return saved
    const detected = detectSystemLocale()
    localStorage.setItem(LOCALE_KEY, detected)
    return detected
  } catch (e) {
    return detectSystemLocale()
  }
}

function initialRemember() {
  try {
    // Default on; only '0' disables it.
    return localStorage.getItem(REMEMBER_KEY) !== '0'
  } catch (e) {
    return true
  }
}

// Module-level singleton so every component shares one preferences instance.
const settings = reactive({
  locale: initialLocale(),
  rememberProgress: initialRemember(),
})

// Persist changes as they happen.
watch(
  () => settings.locale,
  (v) => {
    try {
      localStorage.setItem(LOCALE_KEY, v)
    } catch (e) {
      /* ignore */
    }
  }
)
watch(
  () => settings.rememberProgress,
  (v) => {
    try {
      localStorage.setItem(REMEMBER_KEY, v ? '1' : '0')
    } catch (e) {
      /* ignore */
    }
  }
)

// resolve walks a dotted key path (e.g. "sidebar.title") into a message tree.
function resolve(dict, key) {
  return key.split('.').reduce((o, k) => (o == null ? o : o[k]), dict)
}

// t looks up a key in the active locale, falling back to Chinese and then the
// raw key. {name} placeholders are filled from params. Reading settings.locale
// makes every t() call reactive to language changes.
function t(key, params) {
  let s = resolve(messages[settings.locale], key)
  if (s == null) s = resolve(messages.zh, key)
  if (s == null) return key
  if (params) {
    s = s.replace(/\{(\w+)\}/g, (_, name) =>
      params[name] != null ? String(params[name]) : `{${name}}`
    )
  }
  return s
}

function setLocale(loc) {
  if (messages[loc]) settings.locale = loc
}

function setRememberProgress(on) {
  settings.rememberProgress = !!on
}

export function useSettings() {
  return { settings, t, setLocale, setRememberProgress }
}
