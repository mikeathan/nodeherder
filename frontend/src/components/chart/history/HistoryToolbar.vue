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
    <div class="history-filters">
      <div class="history-chips" role="group" aria-label="Entity">
        <select class="nh-select history-picker" aria-label="Entity" :value="selected"
          @change="$emit('select', ($event.target as HTMLSelectElement).value)">
          <option v-for="name in names" :key="name" :value="name">{{ exposeLabel(name) }}</option>
        </select>
        <button v-for="name in names" :key="name" class="nh-btn history-chip" type="button" :aria-pressed="selected === name"
          @click="$emit('select', name)">{{ exposeLabel(name) }}</button>
      </div>
      <div class="history-chips" role="group" aria-label="Period">
        <select class="nh-select history-picker" aria-label="Period" :value="preset"
          @change="$emit('preset', ($event.target as HTMLSelectElement).value)">
          <option v-for="item in HISTORY_PRESETS" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
        <button v-for="item in HISTORY_PRESETS" :key="item.value" class="nh-btn history-chip" type="button" :aria-pressed="preset === item.value"
          @click="$emit('preset', item.value)">{{ item.label }}</button>
      </div>
    </div>
    <div class="history-actions">
      <div v-if="statistics" role="group" aria-label="History mode" class="history-modes">
        <button class="nh-btn" type="button" :aria-pressed="mode === 'history'" @click="$emit('mode', 'history')">History</button>
        <button class="nh-btn" type="button" :aria-pressed="mode === 'statistics'" @click="$emit('mode', 'statistics')">Statistics</button>
      </div>
      <UiButton variant="ghost" icon="refresh" label="Refresh history" :disabled="loading" @click="$emit('refresh')" />
      <span :title="exportReason || undefined">
        <UiButton class="history-export" icon="export" aria-label="Export CSV" title="Export CSV" :disabled="!!exportReason || loading" @click="$emit('export')"><span class="export-label">Export CSV</span></UiButton>
      </span>
    </div>
    <p v-if="exportReason" class="export-reason">{{ exportReason }}</p>
    <p class="export-reason">{{ mode === 'history' ? 'History CSV uses native sensor values and units.' : 'Statistics CSV uses the displayed values and units.' }}</p>
  </div>
</template>
<style scoped>
  .history-toolbar { display: grid; gap: 0.75rem; min-width: 0; }
  .history-filters { display: contents; }
  .history-picker { display: none; }
  .history-chips, .history-actions, .history-modes { display: flex; flex-wrap: wrap; gap: 0.4rem; align-items: center; }
  button { min-height: 44px; min-width: 44px; }
  button[aria-pressed='true'] { background: var(--nh-accent-soft); border-color: var(--nh-accent); }
  .history-actions { justify-content: flex-end; }
  .history-modes { margin-right: auto; }
  .export-reason { color: var(--nh-text-muted); font-size: 0.75rem; margin: 0; }
  @media (max-width: 719px) {
    .history-toolbar { gap: 0.4rem; }
    .history-filters { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 0.7fr); gap: 0.5rem; }
    .history-chips { min-width: 0; }
    .history-chip { display: none; }
    .history-picker { display: block; width: 100%; height: 44px; font-size: 1rem; text-overflow: ellipsis; }
    .history-actions { flex-wrap: nowrap; gap: 0.25rem; }
    .history-modes { gap: 0; padding: 2px; border: 1px solid var(--nh-border); border-radius: var(--nh-radius-s); background: var(--nh-surface); }
    .history-modes button { border: 0; border-radius: var(--nh-radius-s); padding: 0 0.65rem; font-size: 0.8rem; background: transparent; }
    .history-modes button[aria-pressed='true'] { background: var(--nh-accent-soft); }
    .history-export { width: 44px; min-height: 44px; padding: 0; background: transparent; border-color: transparent; }
    .export-label { display: none; }
  }
</style>
