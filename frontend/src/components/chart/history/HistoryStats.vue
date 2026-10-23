<script setup lang="ts">
  import { computed, defineAsyncComponent } from 'vue';
  import type { Expose } from '@/types/device';
  import type { HistorySeries } from '@/domain/history/types';
  import { numericStats } from '@/domain/history/numeric';
  import { dailyHistoryStats, statisticsDisplaySeries } from '@/domain/history/stats';
  import { statisticsHistoryOptions, type HistoryChartSeries } from './options';
  import { useThemeSettings } from '@/composables/useThemeSettings';
  const HistoryCanvas = defineAsyncComponent(() => import('./HistoryCanvas.vue'));
  const props = defineProps<{ series: HistorySeries; expose: Expose; daily?: boolean }>();
  const { mode, settings } = useThemeSettings();
  const displayed = computed(() => statisticsDisplaySeries(props.series, props.expose));
  const unit = computed(() => displayed.value.unit);
  const summary = computed(() => numericStats(displayed.value.data));
  const buckets = computed(() => props.daily ? dailyHistoryStats(displayed.value) : []);
  const basis = computed(() => displayed.value.data.filter((p) => typeof p.value === 'number').length);
  const format = (value: number | null) => value === null ? 'Unknown' : `${Number(value.toFixed(2))} ${unit.value}`;
  const date = (value: number) => new Date(value).toLocaleDateString([], { month: 'short', day: 'numeric' });
  const plotted = computed<HistoryChartSeries[]>(() => {
    void settings.value;
    void mode.value;
    const css = getComputedStyle(document.documentElement);
    const color = (token: string) => css.getPropertyValue(token).trim();
    return [
      { name: 'Minimum–Maximum', type: 'bar', color: color('--nh-kind-power'), outline: color('--nh-text'), data: buckets.value.map((b) => ({
        x: (b.start + b.end) / 2, y: b.min === null || b.max === null ? null : [b.min, b.max],
      })) },
      { name: 'Average', type: 'line', pointStyle: 'circle', color: color('--nh-kind-temp'), outline: color('--nh-text'),
        data: buckets.value.map((b) => ({ x: (b.start + b.end) / 2, y: b.avg })) },
      { name: 'Last', type: 'line', pointStyle: 'rectRot', color: color('--nh-kind-contact'), outline: color('--nh-text'),
        data: buckets.value.map((b) => ({ x: (b.start + b.end) / 2, y: b.last })) },
    ];
  });
  const options = computed(() => {
    void settings.value;
    const css = getComputedStyle(document.documentElement);
    const color = css.getPropertyValue('--nh-accent').trim();
    return statisticsHistoryOptions({ color, muted: css.getPropertyValue('--nh-text-muted').trim(), grid: css.getPropertyValue('--nh-chart-grid').trim(),
      dark: mode.value === 'dark', compact: false, reducedMotion: window.matchMedia('(prefers-reduced-motion: reduce)').matches },
      buckets.value[0]?.start ?? props.series.from, buckets.value.at(-1)?.end ?? props.series.to, unit.value);
  });
</script>
<template>
  <section aria-label="Sample statistics" class="history-statistics">
    <dl v-if="summary" class="history-summary">
      <div><dt>Range</dt><dd>{{ format(summary.min) }} – {{ format(summary.max) }}</dd></div>
      <div><dt>Sample average</dt><dd>{{ format(summary.avg) }}</dd></div>
      <div><dt>Last sample</dt><dd>{{ format(summary.last) }}</dd></div>
    </dl>
    <p class="sample-basis">Based on {{ basis }} numeric samples{{ series.partial ? ' · partial history' : '' }}</p>
    <template v-if="daily">
      <p class="sample-basis">Daily sample range and readings · axis fits the measured values</p>
      <ul class="statistics-legend" aria-label="Statistics legend">
        <li><span class="range-key" aria-hidden="true">▮</span> Minimum–Maximum range</li>
        <li><span class="average-key" aria-hidden="true">●</span> Average</li>
        <li><span class="last-key" aria-hidden="true">◆</span> Last</li>
      </ul>
      <HistoryCanvas :type="'bar'" :height="300" :options="options" :series="plotted" />
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
</style>
