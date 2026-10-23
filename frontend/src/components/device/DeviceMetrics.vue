<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { useRoute } from 'vue-router';
  import { findDevice } from '@/composables/useHub';
  import { useHistoryQuery } from '@/composables/useHistoryQuery';
  import { historyRange, validHistoryRange } from '@/domain/history/range';
  import { readHistoryPreference, writeHistoryPreference } from '@/domain/history/preferences';
  import { historyCSV, historyExportUnavailable, type HistoryMode } from '@/domain/history/export';
  import { coveredHistoryWindow, statisticsDisplaySeries } from '@/domain/history/stats';
  import { isEnergyCounter } from '@/domain/history/energy';
  import HistoryView from '@/components/chart/history/HistoryView.vue';
  import HistoryToolbar from '@/components/chart/history/HistoryToolbar.vue';
  const props = defineProps<{ id: string }>();
  const route = useRoute();
  const device = computed(() => findDevice(props.id));
  const names = computed(() => Object.values(device.value?.exposes ?? {}).filter((e) => ['numeric', 'binary', 'enum'].includes(String(e.type))).map((e) => e.name).sort());
  const selected = ref('');
  const hours = ref('24');
  const mode = ref<HistoryMode>('history');
  const end = ref(Date.now());
  const customFrom = ref('');
  const customTo = ref('');
  const localInput = (at: number) => new Date(at - new Date(at).getTimezoneOffset() * 60000).toISOString().slice(0, 23);
  let restoredDevice = '';
  let restoredURL: unknown;
  watch([() => props.id, () => route.query.expose, names], ([id, url, values]) => {
    if (!values.length) { selected.value = ''; return; }
    if (id === restoredDevice && url === restoredURL && values.includes(selected.value)) return;
    let preference = { version: 1 as const, preset: '24', expose: values[0], custom: undefined as { from: number; to: number } | undefined };
    try { preference = { ...preference, ...readHistoryPreference(localStorage, id, values) }; } catch { /* Optional storage. */ }
    selected.value = typeof url === 'string' && values.includes(url) ? url : preference.expose;
    hours.value = preference.preset;
    customFrom.value = preference.custom ? localInput(preference.custom.from) : '';
    customTo.value = preference.custom ? localInput(preference.custom.to) : '';
    restoredDevice = id;
    restoredURL = url;
  }, { immediate: true });
  const range = computed(() => historyRange(hours.value, end.value, customFrom.value, customTo.value));
  const valid = computed(() => validHistoryRange(range.value));
  watch([selected, hours, customFrom, customTo], () => {
    if (!selected.value || !valid.value) return;
    try { writeHistoryPreference(localStorage, props.id, { version: 1, preset: hours.value, expose: selected.value,
      ...(hours.value === 'custom' ? { custom: range.value } : {}) }); } catch { /* Optional storage. */ }
  });
  const request = computed(() => valid.value && selected.value ? { deviceId: props.id, exposes: [selected.value], ...range.value } : null);
  const { entry, refresh } = useHistoryQuery(request);
  const series = computed(() => entry.value?.series.find((s) => s.name === selected.value));
  const expose = computed(() => device.value?.exposes[selected.value]);
  const statistics = computed(() => !!series.value && series.value.type === 'numeric' && !isEnergyCounter(series.value));
  watch(series, (value) => { if (value && (value.type !== 'numeric' || isEnergyCounter(value))) mode.value = 'history'; });
  const exportReason = computed(() => entry.value?.status === 'loading' ? 'Loading history…' : historyExportUnavailable(series.value));
  const covered = computed(() => series.value ? coveredHistoryWindow(series.value) : null);
  function reload() { end.value = Date.now(); refresh(); }
  function preset(value: string) { end.value = Date.now(); hours.value = value; }
  function narrow() {
    if (!covered.value || covered.value.to <= covered.value.from) return;
    customFrom.value = localInput(covered.value.from);
    customTo.value = localInput(covered.value.to);
    hours.value = 'custom';
  }
  function download() {
    if (!series.value || exportReason.value) return;
    const exported = mode.value === 'statistics' && expose.value ? statisticsDisplaySeries(series.value, expose.value) : series.value;
    const blob = new Blob([historyCSV(exported, mode.value)], { type: 'text/csv;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `history-${selected.value}-${mode.value}.csv`;
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 0);
  }
</script>
<template>
  <section class="device-history" aria-label="Device history">
    <HistoryToolbar :names="names" :selected="selected" :preset="hours" :mode="mode" :statistics="statistics"
      :loading="entry?.status === 'loading'" :export-reason="exportReason" @select="selected = $event" @preset="preset"
      @mode="mode = $event" @refresh="reload" @export="download" />
    <div v-if="hours === 'custom'" class="custom-history-range">
      <label>From<input v-model="customFrom" class="nh-input" type="datetime-local" step="0.001" /></label>
      <label>To<input v-model="customTo" class="nh-input" type="datetime-local" step="0.001" /></label>
    </div>
    <p class="history-zone">{{ Intl.DateTimeFormat().resolvedOptions().timeZone }}<template v-if="valid"> ·
      {{ new Date(range.from).toLocaleString() }} – {{ new Date(range.to).toLocaleString() }}</template></p>
    <p v-if="!valid" role="status">Choose a valid period of at most 31 days.</p>
    <p v-else-if="entry?.status === 'loading'" role="status">Loading history…</p>
    <p v-else-if="entry?.status === 'error'" role="alert">{{ entry.error }}</p>
    <template v-else-if="series && expose">
      <HistoryView :series="series" :expose="expose" :mode="mode" interactive />
      <button v-if="series.partial && covered && covered.to > covered.from" class="nh-btn" type="button" @click="narrow">Narrow this range</button>
      <p class="history-hint">Focus the chart: + / − zoom · ← / → pan · 0 resets. Export includes the selected period.</p>
    </template>
    <p v-else>No recorded history for this entity.</p>
  </section>
</template>
<style scoped>
  .device-history { width: 100%; min-width: 0; padding: 0.5rem 0; }
  .custom-history-range { display: flex; flex-wrap: wrap; gap: 0.75rem; margin-top: 0.75rem; }
  label { display: flex; flex-direction: column; gap: 0.4rem; flex: 1 1 10rem; font-size: 0.8rem; }
  input { width: 100%; min-height: 44px; }
  .history-zone, .history-hint { font-size: 0.75rem; color: var(--nh-text-muted); overflow-wrap: anywhere; margin: 1rem 0; }
</style>
