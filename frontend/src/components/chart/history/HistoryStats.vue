<script setup lang="ts">
  import { computed, defineAsyncComponent } from 'vue';
  import type { Expose } from '@/types/device';
  import type { HistorySeries } from '@/domain/history/types';
  import { numericStats } from '@/domain/history/numeric';
  import { dailyHistoryStats, statisticsDisplaySeries } from '@/domain/history/stats';
  import { formatWithUnit } from '@/utils/units';
  import { statisticsHistoryOptions, type HistoryChartSeries } from './options';
  import { useChartAppearance } from '@/composables/history/useChartAppearance';
  const HistoryCanvas = defineAsyncComponent(() => import('./HistoryCanvas.vue'));
  const props = defineProps<{ series: HistorySeries; expose: Expose; daily?: boolean }>();
  const appearance = useChartAppearance();
  const displayed = computed(() => statisticsDisplaySeries(props.series, props.expose));
  const unit = computed(() => displayed.value.unit);
  const summary = computed(() => numericStats(displayed.value.data));
  const buckets = computed(() => props.daily ? dailyHistoryStats(displayed.value) : []);
  const basis = computed(() => displayed.value.data.filter((p) => typeof p.value === 'number').length);
  // Units own rounding, spacing and the missing-value placeholder; this must not invent a second
  // sentinel ("Unknown") beside the one every sibling renders ("—").
  const format = (value: number | null) => formatWithUnit(value, unit.value, { digits: 2 });
  const date = (value: number) => new Date(value).toLocaleDateString([], { month: 'short', day: 'numeric' });
  /** Series colours come from the one appearance owner, not from a per-component token read. */
  const plotted = computed<HistoryChartSeries[]>(() => {
    const style = appearance.value;
    // Three measurements share this chart; the legend's three tokens must match the drawn series.
    const color = style.series ?? { power: style.color, temp: style.color, contact: style.color };
    const midpoint = (bucket: { start: number; end: number }) => (bucket.start + bucket.end) / 2;
    return [
      {
        name: 'Minimum–Maximum',
        type: 'bar',
        color: color.power,
        outline: style.text,
        data: buckets.value.map((bucket) => ({
          x: midpoint(bucket),
          y: bucket.min === null || bucket.max === null ? null : [bucket.min, bucket.max],
        })),
      },
      {
        name: 'Average',
        type: 'line',
        pointStyle: 'circle',
        color: color.temp,
        outline: style.text,
        data: buckets.value.map((bucket) => ({ x: midpoint(bucket), y: bucket.avg })),
      },
      {
        name: 'Last',
        type: 'line',
        pointStyle: 'rectRot',
        color: color.contact,
        outline: style.text,
        data: buckets.value.map((bucket) => ({ x: midpoint(bucket), y: bucket.last })),
      },
    ];
  });
  const options = computed(() =>
    statisticsHistoryOptions(
      appearance.value,
      buckets.value[0]?.start ?? props.series.from,
      buckets.value.at(-1)?.end ?? props.series.to,
      unit.value
    )
  );
</script>
<template>
  <section aria-label="Sample statistics" class="history-statistics" :class="{ 'is-daily': daily }">
    <dl v-if="summary" class="history-summary">
      <div><dt>Range</dt><dd>{{ format(summary.min) }} – {{ format(summary.max) }}</dd></div>
      <div><dt>Sample average</dt><dd>{{ format(summary.avg) }}</dd></div>
      <div><dt>Last sample</dt><dd>{{ format(summary.last) }}</dd></div>
    </dl>
    <p class="sample-basis">Based on {{ basis }} numeric samples{{ series.partial ? ' · partial history' : '' }}</p>
    <template v-if="daily">
      <p class="sample-basis statistics-chart-caption">Daily sample range and readings · axis fits the measured values</p>
      <ul class="statistics-legend" aria-label="Statistics legend">
        <li><span class="range-key" aria-hidden="true">▮</span> Minimum–Maximum range</li>
        <li><span class="average-key" aria-hidden="true">●</span> Average</li>
        <li><span class="last-key" aria-hidden="true">◆</span> Last</li>
      </ul>
      <HistoryCanvas class="statistics-canvas" :type="'bar'" :height="300" :options="options" :series="plotted" />
      <div class="statistics-table">
        <table>
          <caption>Daily sample statistics · {{ Intl.DateTimeFormat().resolvedOptions().timeZone }}</caption>
          <thead><tr><th>Day</th><th>Minimum</th><th>Average</th><th>Maximum</th><th>Last</th><th>Samples</th></tr></thead>
          <tbody><tr v-for="bucket in buckets" :key="bucket.start">
            <th>{{ date(bucket.start) }}<small v-if="bucket.start === series.to"> · end-boundary sample</small><small v-else-if="bucket.partial"> · partial day</small></th>
            <td>{{ format(bucket.min) }}</td><td>{{ format(bucket.avg) }}</td><td>{{ format(bucket.max) }}</td>
            <td>{{ format(bucket.last) }}</td><td>{{ bucket.count }}</td>
          </tr></tbody>
        </table>
      </div>
    </template>
  </section>
</template>
<style scoped>
  .sample-basis, caption { color: var(--nh-text-muted); font-size: 0.8rem; padding: 0.75rem 0; text-align: left; }
  .statistics-table { overflow-x: auto; }
  .statistics-legend { display: flex; flex-wrap: wrap; list-style: none; gap: 1rem; padding: 0; font-size: 0.8rem; }
  .range-key { color: var(--nh-kind-power); }
  .average-key { color: var(--nh-kind-temp); }
  .last-key { color: var(--nh-kind-contact); }
  table { width: 100%; border-collapse: collapse; font-size: 0.8rem; }
  td, th { padding: 0.75rem 0.5rem; border-bottom: 1px solid var(--nh-border); text-align: left; white-space: nowrap; }
  @media (max-width: 719px) {
    .is-daily { display: flex; flex-direction: column; --nh-chart-height: clamp(280px, 48dvh, 420px); }
    .is-daily .statistics-chart-caption { display: none; }
    .is-daily .statistics-legend { order: 0; margin: 0.5rem 0; gap: 0.5rem 0.75rem; }
    .is-daily .statistics-canvas { order: 1; }
    .is-daily .history-summary { order: 2; }
    .is-daily > .sample-basis { order: 3; }
    .is-daily .statistics-table { order: 4; }
  }
</style>
