<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed, ref, watch } from 'vue';
  import { DeviceMetricsRequest, DeviceMetrics, DeviceExposeMetrics } from '@/types/metrics.type';
  import Selection from '../input/Selection.vue';
  import { KeyValuePair } from '@/types/types.type';
  import { toUnix } from '@/utils/date.utils';
  import { groupMetricsByType } from '@/contracts/metrics';
  import { ChartComponents } from '@/mixins/useChartComponents';
  import { ChartType, ChartTypes, PeriodType, PeriodOptions, PeriodTypes } from '@/types/chart.type';
  import { getPeriodOffset } from '@/contracts/chart';

  const props = defineProps({
    id: { type: String, required: true },
  });

  const selectePeriod = ref<PeriodType>(PeriodTypes.Today);
  const loading = ref<boolean>(false);
  const hasMetrics = computed(() => Object.keys(groupedMetrics.value).length > 0);

  watch(
    () => props.id,
    () => {
      dateSelected(selectePeriod.value);
    },
    { immediate: true }
  );

  async function dateSelected(value: PeriodType) {
    selectePeriod.value = value;
    const { from, to } = getPeriodOffset(value);
    const request: DeviceMetricsRequest = {
      id: props.id,
      from: toUnix(from),
      to: toUnix(to),
    };
    loading.value = true;
    try {
      await store.dispatch('metrics/query', request);
    } finally {
      loading.value = false;
    }
  }

  const groupedMetrics = computed(() => {
    const results = store.getters['metrics/view'](props.id) as DeviceMetrics;

    return groupMetricsByType(results) as KeyValuePair<DeviceExposeMetrics[]>;
  });

  const orderedMetrics = computed(() => {
    const grouped = groupedMetrics.value;
    const ordered: Array<{ chartType: ChartType; metrics: DeviceExposeMetrics[] }> = [];
    const chartOrder: ChartType[] = [ChartTypes.NumericChart, ChartTypes.BinaryChart];

    chartOrder.forEach((chartType) => {
      const metrics = grouped[chartType];
      if (metrics?.length) {
        ordered.push({ chartType, metrics });
      }
    });

    Object.keys(grouped)
      .filter((chartType) => !chartOrder.includes(chartType as ChartType))
      .sort()
      .forEach((chartType) => {
        const metrics = grouped[chartType];
        if (metrics?.length) {
          ordered.push({ chartType: chartType as ChartType, metrics });
        }
      });

    return ordered;
  });
</script>

<template>
  <div class="sm:col-3 pb-3">
    <Selection
      label="Period:"
      :value="selectePeriod"
      @updated="dateSelected"
      :items="PeriodOptions"
      :disabled="loading">
    </Selection>
  </div>

  <div v-if="loading">
    <p>Loading metrics...</p>
  </div>
  <div v-else-if="!hasMetrics">
    <p>No metrics available</p>
  </div>
  <div v-for="entry in orderedMetrics" :key="entry.chartType">
    <div class="w-full px-0 py-0">
      <component :is="ChartComponents[entry.chartType]" v-bind="{ chartData: entry.metrics }"> </component>
    </div>
  </div>
</template>
