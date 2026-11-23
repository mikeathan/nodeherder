<script setup lang="ts">
  import { computed, defineAsyncComponent } from 'vue';
  import type { Expose } from '@/types/device';
  import type { HistorySeries } from '@/domain/history/types';
  import { displayNumber } from '@/domain/controls';
  import { decimate } from '@/domain/history/numeric';
  import { isEnergyCounter, energyBuckets } from '@/domain/history/energy';
  import { historyDisplayUnit } from '@/domain/history/stats';
  import { exposeDisplayLabel } from '@/domain/exposes';
  import { useChartAppearance } from '@/composables/history/useChartAppearance';
  import { historyOptions } from './options';
  import type { HistoryRange } from '@/domain/history/range';
  import type { HistoryInteraction } from './interactions';
  const HistoryCanvas = defineAsyncComponent(() => import('./HistoryCanvas.vue'));
  const props = defineProps<{ series: HistorySeries; expose: Expose; compact?: boolean; viewRange?: HistoryRange; interaction?: HistoryInteraction }>();
  const appearance = useChartAppearance(props.expose, () => !!props.compact);
  const energy = computed(() => isEnergyCounter(props.series));
  const buckets = computed(() => energyBuckets(props.series.data, props.series.from, props.series.to));
  const chartable = computed(
    () => !energy.value || buckets.value.some((bucket) => bucket.value !== null && bucket.value !== 0)
  );
  const chartNotice = computed(() => {
    if (chartable.value) return '';
    return buckets.value.some((bucket) => bucket.value !== null)
      ? 'No measurable consumption in this period.'
      : 'No complete counter interval in this period.';
  });
  const data = computed(() =>
    energy.value
      ? buckets.value.map((p) => ({ x: p.timestamp, y: p.value }))
      : decimate(props.series.data).map((p) => ({ x: p.timestamp, y: displayValue(p.value) }))
  );
  // One owner for the display unit; this used to be a second copy of the domain's ternary.
  const unit = computed(() => historyDisplayUnit(props.series, props.expose));
  function displayValue(value: unknown): number | null {
    if (typeof value !== 'number') return null;
    return displayNumber(props.expose, value);
  }
  const options = computed(() =>
    historyOptions(
      appearance.value,
      props.viewRange?.from ?? props.series.from,
      props.viewRange?.to ?? props.series.to,
      unit.value,
      ['brightness', 'color_temp', 'battery', 'battpercentage'].includes(props.expose.name),
      energy.value || props.series.name === 'event_count',
      ['power', 'current', 'voltage'].includes(props.expose.name)
    )
  );
</script>
<template>
  <div
    class="history-numeric"
    :class="{ 'is-compact': compact }"
    role="img"
    :aria-label="`${exposeDisplayLabel(expose)} history in ${unit}`">
    <p v-if="chartNotice" class="history-notice">{{ chartNotice }}</p>
    <p v-else-if="!data.length" class="history-notice">
      History has too many gaps to plot. Choose a shorter period in device history.
    </p>
    <HistoryCanvas
      v-else
      :key="energy || series.name === 'event_count' ? 'bar' : 'line'"
      :type="energy || series.name === 'event_count' ? 'bar' : 'line'"
      :height="compact ? 105 : 235"
      :end-marker="!compact && !energy && series.name !== 'event_count'"
      :options="options"
      :interaction="interaction"
      :series="[{ name: expose.name, data }]" />
    <small v-if="energy">Observed consumption · incomplete intervals are partial</small>
  </div>
</template>
<style scoped>
  .history-numeric {
    width: 100%;
    min-width: 0;
  }
  .history-notice {
    color: var(--nh-text-muted);
    font-size: 0.85rem;
  }
  small {
    display: block;
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
</style>
