<script setup lang="ts">
  import { HISTORY_PRESETS } from '@/domain/history/range';
  import { exposeLabel } from '@/domain/exposes';
  import type { HistoryMode } from '@/domain/history/export';
  import UiButton from '@/components/ui/UiButton.vue';
  defineProps<{ names: string[]; selected: string; preset: string; mode: HistoryMode; statistics: boolean; loading: boolean; exportReason: string }>();
  defineEmits<{ select: [name: string]; preset: [value: string]; mode: [value: HistoryMode]; refresh: []; export: [] }>();
</script>
<template>
  <div class="history-toolbar">
    <div class="history-chips" role="group" aria-label="Entity">
      <button v-for="name in names" :key="name" class="nh-btn" type="button" :aria-pressed="selected === name"
        @click="$emit('select', name)">{{ exposeLabel(name) }}</button>
    </div>
    <div class="history-chips" role="group" aria-label="Period">
      <button v-for="item in HISTORY_PRESETS" :key="item.value" class="nh-btn" type="button" :aria-pressed="preset === item.value"
        @click="$emit('preset', item.value)">{{ item.label }}</button>
    </div>
    <div class="history-actions">
      <div v-if="statistics" role="group" aria-label="History mode" class="history-modes">
        <button class="nh-btn" type="button" :aria-pressed="mode === 'history'" @click="$emit('mode', 'history')">History</button>
        <button class="nh-btn" type="button" :aria-pressed="mode === 'statistics'" @click="$emit('mode', 'statistics')">Statistics</button>
      </div>
      <UiButton variant="ghost" icon="refresh" label="Refresh history" :disabled="loading" @click="$emit('refresh')" />
      <span :title="exportReason || undefined">
        <UiButton icon="export" :disabled="!!exportReason || loading" @click="$emit('export')">Export CSV</UiButton>
      </span>
    </div>
    <p v-if="exportReason" class="export-reason">{{ exportReason }}</p>
    <p class="export-reason">{{ mode === 'history' ? 'History CSV uses native sensor values and units.' : 'Statistics CSV uses the displayed values and units.' }}</p>
  </div>
</template>
<style scoped>
  .history-toolbar { display: grid; gap: 0.75rem; min-width: 0; }
  .history-chips, .history-actions, .history-modes { display: flex; flex-wrap: wrap; gap: 0.4rem; align-items: center; }
  button { min-height: 44px; min-width: 44px; }
  button[aria-pressed='true'] { background: var(--nh-accent-soft); border-color: var(--nh-accent); }
  .history-actions { justify-content: flex-end; }
  .history-modes { margin-right: auto; }
  .export-reason { color: var(--nh-text-muted); font-size: 0.75rem; margin: 0; }
  @media (max-width: 719px) {
    .history-chips { flex-wrap: nowrap; overflow-x: auto; padding-bottom: 0.25rem; }
    .history-chips button { flex: 0 0 auto; }
  }
</style>
