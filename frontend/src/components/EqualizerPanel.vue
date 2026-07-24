<script setup>
import { usePlayer, EQ_FREQUENCIES, EQ_PRESETS } from '../composables/usePlayer'

defineEmits(['close'])

const { state, setEqEnabled, setEqGain, applyEqPreset } = usePlayer()

const presetKeys = Object.keys(EQ_PRESETS)

function freqLabel(f) {
  return f >= 1000 ? f / 1000 + 'k' : String(f)
}

function onGainInput(i, e) {
  setEqGain(i, Number(e.target.value))
}
</script>

<template>
  <div class="panel eq-panel">
    <div class="panel-header">
      <div class="panel-title">
        <span>均衡器</span>
        <label class="switch">
          <input type="checkbox" :checked="state.eqEnabled" @change="setEqEnabled($event.target.checked)" />
          <span class="slider-toggle"></span>
        </label>
      </div>
      <button class="close-btn" aria-label="关闭" @click="$emit('close')">
        <svg width="14" height="14" viewBox="0 0 12 12">
          <path d="M1 1l10 10M11 1L1 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
      </button>
    </div>

    <div class="presets">
      <button
        v-for="key in presetKeys"
        :key="key"
        class="preset-btn"
        :class="{ active: state.eqPreset === key }"
        @click="applyEqPreset(key)"
      >
        {{ EQ_PRESETS[key].label }}
      </button>
    </div>

    <div class="eq-bands" :class="{ disabled: !state.eqEnabled }">
      <div v-for="(f, i) in EQ_FREQUENCIES" :key="f" class="eq-band">
        <span class="gain-value">{{ state.eqGains[i] > 0 ? '+' : '' }}{{ state.eqGains[i] }}</span>
        <input
          class="eq-slider"
          type="range"
          min="-12"
          max="12"
          step="1"
          :value="state.eqGains[i]"
          :disabled="!state.eqEnabled"
          @input="onGainInput(i, $event)"
        />
        <span class="freq-label">{{ freqLabel(f) }}</span>
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
  margin-bottom: 18px;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 12px;
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

/* Toggle switch */
.switch {
  position: relative;
  display: inline-block;
  width: 34px;
  height: 18px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider-toggle {
  position: absolute;
  inset: 0;
  background: var(--progress-track);
  border-radius: 10px;
  transition: background 0.2s;
}

.slider-toggle::before {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  left: 2px;
  top: 2px;
  background: var(--fg);
  border-radius: 50%;
  transition: transform 0.2s;
}

.switch input:checked + .slider-toggle {
  background: var(--primary);
}

.switch input:checked + .slider-toggle::before {
  transform: translateX(16px);
  background: var(--bg);
}

.presets {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 24px;
}

.preset-btn {
  font-size: 12px;
  padding: 6px 12px;
  color: var(--text-secondary);
  background: color-mix(in srgb, var(--seed-fg) 5%, transparent);
  border: 1px solid var(--border);
  border-radius: calc(var(--radius) * 0.6);
  cursor: pointer;
  transition: all 0.15s;
}

.preset-btn:hover {
  color: var(--fg);
  border-color: color-mix(in srgb, var(--seed-primary) 40%, transparent);
}

.preset-btn.active {
  color: var(--bg);
  background: var(--primary);
  border-color: var(--primary);
}

.eq-bands {
  flex: 1;
  display: flex;
  justify-content: space-between;
  align-items: stretch;
  gap: 8px;
  transition: opacity 0.2s;
}

.eq-bands.disabled {
  opacity: 0.4;
}

.eq-band {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.gain-value {
  font-size: 11px;
  color: var(--text-tertiary);
  font-variant-numeric: tabular-nums;
}

.freq-label {
  font-size: 11px;
  color: var(--text-secondary);
}

/* Vertical range slider (writing-mode approach works in modern WebView2) */
.eq-slider {
  writing-mode: vertical-lr;
  direction: rtl;
  width: 6px;
  flex: 1;
  min-height: 160px;
  cursor: pointer;
  accent-color: var(--primary);
}
</style>
