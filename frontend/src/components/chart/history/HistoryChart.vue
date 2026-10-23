<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import type { HistorySeries } from '@/domain/history/types';
  import type { Expose } from '@/types/device';
  import { clampHistoryRange, panHistoryRange, zoomHistoryRange, type HistoryRange } from '@/domain/history/range';
  import { formatExposeValue, exposeLabel } from '@/domain/exposes';
  import NumericHistory from './NumericHistory.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import type { HistoryInteraction } from './interactions';
  const mobileTools = ref<HTMLDetailsElement | null>(null);
  function closeTools() {
    if (!mobileTools.value) return;
    mobileTools.value.open = false;
    mobileTools.value.querySelector('summary')?.focus();
  }
  const props = defineProps<{ series: HistorySeries; expose: Expose; compact?: boolean }>();
  const bounds = computed(() => ({ from: props.series.from, to: props.series.to }));
  const range = ref<HistoryRange>({ ...bounds.value });
  const readout = ref('');
  watch(() => props.series, () => { range.value = { ...bounds.value }; readout.value = ''; });
  const rangeLabel = computed(() => `${new Date(range.value.from).toLocaleString()} – ${new Date(range.value.to).toLocaleString()}`);
  function change(next: HistoryRange) { range.value = clampHistoryRange(next, bounds.value); readout.value = ''; }
  function scrub(at: number | null) {
    if (at === null) { readout.value = ''; return; }
    const points = props.series.data;
    let low = 0, high = points.length;
    while (low < high) { const mid = (low + high) >>> 1; if (points[mid].timestamp < at) low = mid + 1; else high = mid; }
    const right = points[Math.min(low, points.length - 1)], left = points[Math.max(0, low - 1)];
    const point = right && left ? (Math.abs(right.timestamp - at) < Math.abs(left.timestamp - at) ? right : left) : right;
    readout.value = point ? `${new Date(point.timestamp).toLocaleString()} · ${formatExposeValue(props.expose, point.value)}` : 'No recorded sample';
  }
  const controls = [
    { label: 'Pan earlier', text: '←', run: () => change(panHistoryRange(range.value, bounds.value, -1)) },
    { label: 'Zoom out', text: '−', run: () => change(zoomHistoryRange(range.value, bounds.value, 2)) },
    { label: 'Zoom in', text: '+', run: () => change(zoomHistoryRange(range.value, bounds.value, 0.5)) },
    { label: 'Reset', text: 'Reset', run: () => change(bounds.value) },
    { label: 'Pan later', text: '→', run: () => change(panHistoryRange(range.value, bounds.value, 1)) },
  ];
  const interaction: HistoryInteraction = { range: () => range.value, bounds: () => bounds.value, change, readout: scrub };
</script>
<template>
  <div data-history-chart :class="{ 'is-compact': compact }" class="history-chart" tabindex="0"
    role="group" :aria-label="`${exposeLabel(expose.name)} interactive history. ${rangeLabel}. Plus and minus zoom, arrows pan, zero resets.`"
    :data-from="range.from" :data-to="range.to">
    <div class="chart-actions desktop-chart-actions" aria-label="Chart controls">
      <button v-for="control in controls" :key="control.label" class="nh-btn" type="button" :aria-label="control.label" @click="control.run">{{ control.text }}</button>
    </div>
    <NumericHistory :series="series" :expose="expose" :compact="compact" :view-range="range" :interaction="interaction" />
    <p class="chart-range">{{ rangeLabel }}</p>
    <div class="chart-footer">
      <p class="chart-readout" role="status" aria-live="off"><template v-if="readout">{{ readout }}</template><template v-else><span class="desktop-hint">Drag to inspect samples · pinch or wheel to zoom</span><span class="mobile-hint">Touch to inspect · pinch to zoom</span></template></p>
      <details ref="mobileTools" class="mobile-chart-tools" @keydown.esc.stop.prevent="closeTools">
        <summary>Chart tools <UiIcon name="chevronDown" /></summary>
        <div class="chart-actions" aria-label="Chart controls">
          <button v-for="control in controls" :key="control.label" class="nh-btn" type="button" :aria-label="control.label" @click="control.run">{{ control.text }}</button>
        </div>
        <p class="tool-range">Visible range: {{ rangeLabel }}</p>
      </details>
    </div>
  </div>
</template>
<style scoped>
  .history-chart { min-width: 0; border-radius: var(--nh-radius-m); }
  .history-chart:not(.is-compact) { --nh-chart-height: 380px; }
  @media (max-width: 1023px) { .history-chart:not(.is-compact) { --nh-chart-height: 300px; } }
  @media (max-width: 719px) { .history-chart:not(.is-compact) { --nh-chart-height: clamp(280px, 48dvh, 420px); } }
  .history-chart:focus-visible { outline: 3px solid var(--nh-focus); outline-offset: 3px; }
  .mobile-chart-tools, .mobile-hint { display: none; }
  .chart-actions { display: flex; justify-content: flex-end; gap: 0.35rem; flex-wrap: wrap; }
  .chart-actions button { min-width: 44px; min-height: 44px; }
  .chart-range, .chart-readout { font-size: 0.75rem; color: var(--nh-text-muted); overflow-wrap: anywhere; margin: 0.35rem 0; }
  @media (max-width: 719px) {
    .desktop-chart-actions, .chart-range, .desktop-hint { display: none; }
    .mobile-hint { display: inline; }
    .chart-footer { display: flex; flex-wrap: wrap; align-items: flex-start; gap: 0.5rem; }
    .chart-readout { flex: 1; min-width: 0; padding-top: 0.6rem; }
    .mobile-chart-tools { display: block; }
    .mobile-chart-tools[open] { flex-basis: 100%; }
    summary { display: flex; align-items: center; justify-content: flex-end; min-height: 44px; min-width: 44px;
      gap: 0.25rem; list-style: none; cursor: pointer; font-size: 0.75rem; color: var(--nh-text-muted); }
    .mobile-chart-tools[open] summary :deep(.nh-ic) { transform: rotate(180deg); }
    summary::-webkit-details-marker { display: none; }
    summary:focus-visible { outline: 2px solid var(--nh-focus); outline-offset: 2px; }
    .tool-range { font-size: 0.75rem; color: var(--nh-text-muted); overflow-wrap: anywhere; }
  }
</style>
