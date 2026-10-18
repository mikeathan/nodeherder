<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import type { Expose } from '@/types/device';
  import type { HistorySeries } from '@/domain/history/types';
  import { isEnergyCounter, energyBuckets, energySummary } from '@/domain/history/energy';
  import { eventCounts } from '@/domain/history/events';
  import { numericStats } from '@/domain/history/numeric';
  import { exposeKind, formatExposeValue } from '@/domain/exposes';
  import NumericHistory from './NumericHistory.vue';
  import StateHistory from './StateHistory.vue';
  const props = defineProps<{ series: HistorySeries; expose: Expose; compact?: boolean }>();
  const isEnergy = computed(() => isEnergyCounter(props.series));
  const consumption = computed(() =>
    energySummary(energyBuckets(props.series.data, props.series.from, props.series.to))
  );
  const stats = computed(() => numericStats(props.series.data));
  const recent = computed(() => props.series.data.slice(-6).reverse());
  const samplePage = ref(0);
  const pageCount = computed(() => Math.ceil(props.series.data.length / 100));
  const samples = computed(() =>
    [...props.series.data].reverse().slice(samplePage.value * 100, (samplePage.value + 1) * 100)
  );
  watch(
    () => props.series,
    () => {
      samplePage.value = 0;
    }
  );
  const format = (value: unknown) =>
    formatExposeValue({ ...props.expose, unit: props.series.unit || props.expose.unit }, value);
  const time = (value: number) =>
    new Date(value).toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  const counts = computed(() => eventCounts(props.series));
</script>
<template>
  <section
    class="history-view"
    :class="`kind-${exposeKind(expose.name)}`"
    :style="{ '--history-color': `var(--nh-kind-${exposeKind(expose.name)})` }">
    <p v-if="series.partial" class="history-notice" role="status">
      Partial history · narrow the period for complete results.
    </p>
    <p v-if="!series.data.length" class="history-notice">No recorded history for this period.</p>
    <template v-else>
      <NumericHistory v-if="series.type === 'numeric'" :series="series" :expose="expose" :compact="compact" />
      <StateHistory v-else-if="series.type === 'binary'" :series="series" :expose="expose" :compact="compact" />
      <NumericHistory
        v-else
        :series="counts"
        :expose="{ ...expose, name: 'event_count', type: 'numeric', unit: 'events' }"
        :compact="compact" />
      <dl v-if="stats && !compact && series.type === 'numeric' && !isEnergy" class="history-summary">
        <div>
          <dt>Range</dt>
          <dd>{{ format(stats.min) }} – {{ format(stats.max) }}</dd>
        </div>
        <div>
          <dt>Sample average</dt>
          <dd>{{ format(stats.avg) }}</dd>
        </div>
      </dl>
      <dl v-if="isEnergy && !compact" class="history-summary">
        <div>
          <dt>Observed consumption</dt>
          <dd>{{ format(consumption.observed) }}</dd>
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
          ><span>{{ format(point.value) }}</span>
        </div>
      </div>
      <details v-if="!compact" class="history-samples">
        <summary>Inspect recorded samples</summary>
        <p>
          Page {{ samplePage + 1 }} of {{ pageCount }} · {{ series.data.length }} recorded samples ·
          {{ Intl.DateTimeFormat().resolvedOptions().timeZone }}
        </p>
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Value</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(point, i) in samples" :key="i">
              <td :title="new Date(point.timestamp).toISOString()">{{ new Date(point.timestamp).toLocaleString() }}</td>
              <td>{{ format(point.value) }}</td>
            </tr>
          </tbody>
        </table>
        <div class="sample-pages">
          <button type="button" class="nh-btn" :disabled="samplePage === 0" @click="samplePage--">Newer samples</button>
          <button type="button" class="nh-btn" :disabled="samplePage + 1 >= pageCount" @click="samplePage++">
            Older samples
          </button>
        </div>
      </details>
    </template>
  </section>
</template>
<style>
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
