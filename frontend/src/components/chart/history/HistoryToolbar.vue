<script setup lang="ts">
  import { ref } from 'vue';
  import { HISTORY_PRESETS } from '@/domain/history/range';
  import { exposeLabel } from '@/domain/exposes';
  import type { HistoryMode } from '@/domain/history/export';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';
  defineProps<{ names: string[]; selected: string; preset: string; mode: HistoryMode; statistics: boolean; loading: boolean; exportReason: string; zone: string; rangeLabel: string }>();
  const emit = defineEmits<{ select: [name: string]; preset: [value: string]; mode: [value: HistoryMode]; refresh: []; export: [] }>();
  const optionsOpen = ref(false);
  function changeMode(event: Event) {
    const value = (event.target as HTMLSelectElement).value;
    if (value === 'history' || value === 'statistics') emit('mode', value);
    optionsOpen.value = false;
  }
  function refresh() { emit('refresh'); optionsOpen.value = false; }
  function download() { emit('export'); optionsOpen.value = false; }
</script>
<template>
  <div class="history-toolbar">
    <div class="history-filters">
      <div class="history-chips" role="group" aria-label="Entity">
        <span class="history-field-label" aria-hidden="true">Entity</span>
        <select class="nh-select history-picker" aria-label="Entity" :value="selected"
          @change="$emit('select', ($event.target as HTMLSelectElement).value)">
          <option v-for="name in names" :key="name" :value="name">{{ exposeLabel(name) }}</option>
        </select>
      </div>
      <div class="history-chips" role="group" aria-label="Period">
        <span class="history-field-label" aria-hidden="true">Period</span>
        <select class="nh-select history-picker" aria-label="Period" :value="preset"
          @change="$emit('preset', ($event.target as HTMLSelectElement).value)">
          <option v-for="item in HISTORY_PRESETS" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
      </div>
    </div>
    <UiButton class="mobile-history-options" variant="ghost" icon="more" label="History options"
      :aria-description="`Current view: ${mode === 'history' ? 'History' : 'Statistics'}`" @click="optionsOpen = true" />
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
    <p class="history-unit-help">{{ mode === 'history' ? 'History CSV uses native sensor values and units.' : 'Statistics CSV uses the displayed values and units.' }}</p>
  </div>
  <UiDialog :open="optionsOpen" title="History options" @close="optionsOpen = false">
    <div class="history-settings">
      <p>{{ exposeLabel(selected) }}</p>
      <label v-if="statistics">View
        <select class="nh-select" :value="mode" @change="changeMode">
          <option value="history">History</option><option value="statistics">Statistics</option>
        </select>
      </label>
      <p class="settings-range">{{ zone }}<template v-if="rangeLabel"> · {{ rangeLabel }}</template></p>
      <div class="settings-actions">
        <UiButton icon="refresh" :disabled="loading" @click="refresh">Refresh history</UiButton>
        <UiButton icon="export" :disabled="!!exportReason || loading" @click="download">Export CSV</UiButton>
      </div>
      <p v-if="exportReason" class="export-reason">{{ exportReason }}</p>
      <p class="history-unit-help">{{ mode === 'history' ? 'History CSV uses native sensor values and units.' : 'Statistics CSV uses the displayed values and units.' }}</p>
    </div>
  </UiDialog>
</template>
<style scoped>
  .history-toolbar { display: grid; gap: 0.75rem; min-width: 0; }
  .history-filters { display: contents; }
  .history-picker, .history-field-label, .mobile-history-options { display: none; }
  .history-chips, .history-actions, .history-modes { display: flex; flex-wrap: wrap; gap: 0.4rem; align-items: center; }
  button { min-height: 44px; min-width: 44px; }
  button[aria-pressed='true'] { background: var(--nh-accent-soft); border-color: var(--nh-accent); }
  .history-actions { justify-content: flex-end; }
  .history-modes { margin-right: auto; }
  .export-reason, .history-unit-help, .settings-range { color: var(--nh-text-muted); font-size: 0.75rem; margin: 0; }
  .history-settings { display: grid; gap: 1rem; }
  .history-settings label { display: grid; gap: 0.5rem; }
  .history-settings select { width: 100%; height: 44px; font-size: 1rem; }
  .settings-actions { display: flex; flex-wrap: wrap; gap: 0.5rem; }
  .settings-actions :deep(button) { min-height: 44px; }
  @media (min-width: 720px) {
    .history-toolbar { grid-template-columns: minmax(11rem, 20rem) 8.5rem minmax(max-content, 1fr);
      column-gap: 0.75rem; row-gap: 0.4rem; align-items: end; }
    .history-chips { display: grid; gap: 0.35rem; min-width: 0; }
    .history-field-label { display: block; color: var(--nh-text-muted); font-size: 0.75rem; }
    .history-picker { display: block; width: 100%; height: 44px; }
    .history-actions { flex-wrap: nowrap; min-width: 0; }
    .history-modes { flex-wrap: nowrap; gap: 0; }
    .history-modes button { border: 0; border-bottom: 2px solid transparent; border-radius: 0;
      padding: 0 0.75rem; background: transparent; }
    .history-modes button[aria-pressed='true'] { background: transparent; color: var(--nh-accent); border-bottom-color: var(--nh-accent); }
    .history-export { background: transparent; border-color: transparent; padding-inline: 0.5rem; }
    .history-toolbar > .export-reason, .history-toolbar > .history-unit-help { grid-column: 1 / -1; }
  }
  @media (max-width: 719px) {
    .history-toolbar { grid-template-columns: minmax(0, 1fr) 6.75rem 44px; gap: 0.375rem; align-items: center; }
    .history-filters { display: contents; }
    .history-chips { min-width: 0; }
    .history-picker { display: block; width: 100%; height: 44px; font-size: 1rem; text-overflow: ellipsis; }
    .history-chips:last-child .history-picker { padding-inline: 0.3rem; }
    .history-actions, .history-toolbar > .history-unit-help { display: none; }
    .mobile-history-options { display: inline-flex; width: 44px; min-height: 44px; }
    .history-toolbar > .export-reason { grid-column: 1 / -1; }
  }
</style>
