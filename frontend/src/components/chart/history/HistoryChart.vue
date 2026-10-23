<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import type { HistorySeries } from '@/domain/history/types';
  import type { Expose } from '@/types/device';
  import { clampHistoryRange, panHistoryRange, zoomHistoryRange, type HistoryRange } from '@/domain/history/range';
  import { formatExposeValue, exposeLabel } from '@/domain/exposes';
  import NumericHistory from './NumericHistory.vue';
  import type { HistoryInteraction } from './interactions';
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
  const interaction: HistoryInteraction = { range: () => range.value, bounds: () => bounds.value, change, readout: scrub };
</script>
<template>
  <div data-history-chart :class="{ 'is-compact': compact }" class="history-chart" tabindex="0"
    role="group" :aria-label="`${exposeLabel(expose.name)} interactive history. ${rangeLabel}. Plus and minus zoom, arrows pan, zero resets.`"
    :data-from="range.from" :data-to="range.to">
    <div class="chart-actions" aria-label="Chart controls">
      <button class="nh-btn" type="button" aria-label="Pan earlier" @click="change(panHistoryRange(range, bounds, -1))">←</button>
      <button class="nh-btn" type="button" aria-label="Zoom out" @click="change(zoomHistoryRange(range, bounds, 2))">−</button>
      <button class="nh-btn" type="button" aria-label="Zoom in" @click="change(zoomHistoryRange(range, bounds, 0.5))">+</button>
      <button class="nh-btn" type="button" @click="change(bounds)">Reset</button>
      <button class="nh-btn" type="button" aria-label="Pan later" @click="change(panHistoryRange(range, bounds, 1))">→</button>
    </div>
    <NumericHistory :series="series" :expose="expose" :compact="compact" :view-range="range" :interaction="interaction" />
    <p class="chart-range">{{ rangeLabel }}</p>
    <p class="chart-readout" role="status" aria-live="off">{{ readout || 'Drag to inspect samples · pinch or wheel to zoom' }}</p>
  </div>
</template>
<style scoped>
  .history-chart { min-width: 0; border-radius: var(--nh-radius-m); }
  .history-chart:not(.is-compact) { --nh-chart-height: 380px; }
  @media (max-width: 1023px) { .history-chart:not(.is-compact) { --nh-chart-height: 300px; } }
  @media (max-width: 719px) { .history-chart:not(.is-compact) { --nh-chart-height: 240px; } }
  .history-chart:focus-visible { outline: 3px solid var(--nh-focus); outline-offset: 3px; }
  .chart-actions { display: flex; justify-content: flex-end; gap: 0.35rem; flex-wrap: wrap; }
  .chart-actions button { min-width: 44px; min-height: 44px; }
  .chart-range, .chart-readout { font-size: 0.75rem; color: var(--nh-text-muted); overflow-wrap: anywhere; margin: 0.35rem 0; }
</style>
