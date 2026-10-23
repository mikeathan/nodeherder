<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import type { HistorySeries } from '@/domain/history/types';
  import type { Expose } from '@/types/device';
  import { formatExposeValue } from '@/domain/exposes';
  const props = defineProps<{ series: HistorySeries; expose: Expose }>();
  const samplePage = ref(0);
  const pageCount = computed(() => Math.ceil(props.series.data.length / 100));
  const samples = computed(() => [...props.series.data].reverse().slice(samplePage.value * 100, (samplePage.value + 1) * 100));
  watch(() => props.series, () => { samplePage.value = 0; });
  const format = (value: unknown) => formatExposeValue({ ...props.expose, unit: props.series.unit || props.expose.unit }, value);
</script>
<template>
      <details class="history-samples">
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
<style scoped>
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
