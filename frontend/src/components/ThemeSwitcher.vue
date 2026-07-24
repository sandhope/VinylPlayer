<script setup>
import { useTheme } from '../composables/useTheme'

const { current, themes, apply } = useTheme()
</script>

<template>
  <div class="theme-switcher">
    <div
      v-for="t in themes"
      :key="t.id"
      class="theme-dot"
      :class="[t.cls, { active: current === t.id }]"
      :title="t.label"
      role="button"
      :aria-label="t.label"
      @click="apply(t.id)"
    ></div>
  </div>
</template>

<style scoped>
.theme-switcher {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 10;
  display: flex;
  gap: 6px;
  background: var(--surface);
  padding: 6px;
  border-radius: calc(var(--radius) * 1.5);
  border: 1px solid var(--border);
}

.theme-dot {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  cursor: pointer;
  transition: border-color 0.2s, transform 0.15s;
  position: relative;
}

.theme-dot:hover {
  transform: scale(1.15);
}

.theme-dot.active {
  border: 2px solid var(--fg);
}

.theme-dot.retro {
  background: linear-gradient(135deg, #d4874e, #c9a96e);
}

.theme-dot.dark {
  background: linear-gradient(135deg, #2dd4a8, #0f2027);
}

.theme-dot.light {
  background: linear-gradient(135deg, #f8fafc, #94a3b8);
}

.theme-dot.minimal {
  background: linear-gradient(135deg, #fafafa, #171717);
}
</style>
