<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed, onMounted, ref, watch } from 'vue';
  import { DeviceMetricsRequest, DeviceMetrics, DeviceExposeMetrics } from '@/types/metrics.type';
  import Selection from '../input/Selection.vue';
  import { KeyValuePair } from '@/types/types.type';
  import { toUnix } from '@/utils/date.utils';
  import { groupMetricsByType } from '@/contracts/metrics';
  import { ChartComponents } from '@/mixins/useChartComponents';
  import { PeriodType, PeriodOptions, PeriodTypes } from '@/types/chart.type';
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
</script>

<template>
  <div class="sm:col-3">
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
  <div v-for="(metrics, chartType) in groupedMetrics">
    <component :is="ChartComponents[chartType]" v-bind="{ chartData: metrics }"> </component>
  </div>
</template>
