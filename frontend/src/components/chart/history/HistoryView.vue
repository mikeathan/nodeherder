<script setup lang="ts">
  import { computed } from 'vue';
  import type { Expose } from '@/types/device';
  import type { HistorySeries } from '@/domain/history/types';
  import { isEnergyCounter, energyBuckets, energySummary } from '@/domain/history/energy';
  import { eventCounts } from '@/domain/history/events';
  import HistoryChart from './HistoryChart.vue';
  import HistoryStats from './HistoryStats.vue';
  import HistorySamples from './HistorySamples.vue';
  import type { HistoryMode } from '@/domain/history/export';
  import { coveredHistoryWindow } from '@/domain/history/stats';
  import { exposeKind, exposeValueColor, formatExposeValue } from '@/domain/exposes';
  import { formatHistoryMoment } from '@/domain/history/format';
  import NumericHistory from './NumericHistory.vue';
  import StateHistory from './StateHistory.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  const props = defineProps<{ series: HistorySeries; expose: Expose; compact?: boolean; mode?: HistoryMode; interactive?: boolean }>();
  const isEnergy = computed(() => isEnergyCounter(props.series));
  const buckets = computed(() =>
    isEnergy.value ? energyBuckets(props.series.data, props.series.from, props.series.to) : []
  );
  const consumption = computed(() => energySummary(buckets.value));
  const recent = computed(() => props.series.data.slice(-6).reverse());
  const covered = computed(() => coveredHistoryWindow(props.series));
  const format = (value: unknown) =>
    formatExposeValue({ ...props.expose, unit: props.series.unit || props.expose.unit }, value);
  const time = (value: number) => formatHistoryMoment(value);
  const counts = computed(() => eventCounts(props.series));
  const reportColor = (value: unknown) => exposeValueColor(props.expose, value);
</script>
<template>
  <section
    class="history-view"
    :class="`kind-${exposeKind(expose.name)}`">
    <p v-if="series.partial" class="history-notice" role="status">
      Partial history · narrow the period for complete results.
    </p>
    <p v-if="series.partial && covered" class="history-notice">Covered samples: {{ new Date(covered.from).toLocaleString() }} – {{ new Date(covered.to).toLocaleString() }}. Newer samples may be missing.</p>
    <div v-if="!series.data.length" class="history-empty">
      <UiIcon name="chart" />
      <p>No recorded history for this period.</p>
      <small>Charts appear when this device reports and metrics are recorded.</small>
    </div>
    <template v-else>
      <HistoryStats v-if="mode === 'statistics' && series.type === 'numeric' && !isEnergy" :series="series" :expose="expose" daily />
      <HistoryChart v-else-if="interactive && series.type === 'numeric'" :series="series" :expose="expose" :compact="compact" />
      <NumericHistory v-else-if="series.type === 'numeric'" :series="series" :expose="expose" :compact="compact" />
      <StateHistory v-else-if="series.type === 'binary'" :series="series" :expose="expose" :compact="compact" />
      <NumericHistory
        v-else
        :series="counts"
        :expose="{ ...expose, name: 'event_count', type: 'numeric', unit: 'events' }"
        :compact="compact" />
      <HistoryStats v-if="!compact && series.type === 'numeric' && !isEnergy && mode !== 'statistics'" :series="series" :expose="expose" />
      <dl v-if="isEnergy" class="history-summary">
        <div>
          <dt>Observed consumption</dt>
          <dd>{{ consumption.observed === null ? 'Not enough data' : format(consumption.observed) }}</dd>
        </div>
        <div>
          <dt>Complete intervals</dt>
          <dd>{{ consumption.complete }} / {{ consumption.count }}</dd>
        </div>
      </dl>
      <div v-if="!compact && series.type !== 'numeric'" class="history-recent">
        <h3>{{ series.type === 'enum' ? 'Recent events' : 'Recent reports' }}</h3>
        <div v-for="(point, i) in recent" :key="i">
          <time>{{ time(point.timestamp) }}</time
          ><span :style="{ '--report-color': reportColor(point.value) }">{{ format(point.value) }}</span>
        </div>
      </div>
      <HistorySamples v-if="!compact" :series="series" :expose="expose" />
    </template>
  </section>
</template>
<style>
  /* Shared with StateHistory: both render the same summary rows. */
  .history-summary {
    margin: 1rem 0 0;
  }
  .history-summary > div {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.9rem 0;
    border-top: 1px solid var(--nh-border);
  }
  .history-summary dt {
    color: var(--nh-text-muted);
  }
  .history-summary dd {
    margin: 0;
    font-weight: 600;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }
</style>
<style scoped>
  .history-view {
    width: 100%;
    min-width: 0;
  }
  .history-notice {
    color: var(--nh-text-muted);
    font-size: 0.85rem;
  }
  .history-empty {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 0.5rem 0;
    color: var(--nh-text-muted);
  }
  .history-empty :deep(.nh-ic) {
    width: 1.5rem;
    height: 1.5rem;
  }
  .history-empty p,
  .history-empty small {
    margin: 0;
  }
  .history-empty p {
    font-size: 0.85rem;
  }
  .history-empty small {
    font-size: 0.75rem;
  }
  .history-recent {
    margin-top: 1.5rem;
  }
  h3 {
    font-size: 0.9rem;
    margin-bottom: 0.5rem;
  }
  .history-recent > div {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.7rem 0;
    border-bottom: 1px solid var(--nh-border);
    font-size: 0.85rem;
  }
  time {
    color: var(--nh-text-muted);
  }
  .history-recent span::before {
    content: '';
    display: inline-block;
    width: 0.5rem;
    height: 0.5rem;
    border-radius: 50%;
    margin-right: 0.5rem;
    background: var(--report-color);
  }
  .history-samples {
    margin-top: 1rem;
    font-size: 0.8rem;
  }
  summary {
    cursor: pointer;
    padding: 0.5rem 0;
    color: var(--nh-text-muted);
  }
  .sample-pages {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    padding-top: 0.75rem;
  }
  table {
    width: 100%;
    border-collapse: collapse;
  }
  td,
  th {
    text-align: left;
    padding: 0.5rem 0.2rem;
    border-bottom: 1px solid var(--nh-border);
  }
</style>
