<script setup lang="ts">
import { store } from '../../store/index';
import { computed, onMounted, ref, watch } from 'vue';
import {
  DeviceMetricsRequest,
  DeviceMetrics,
  DeviceExposeMetrics,
} from '@/types/metrics.type';
import Selection from '../input/Selection.vue';
import { KeyValuePair } from '@/types/types';
import { toUnix } from '@/utils/date.utils';
import { groupMetricsByType } from '@/contracts/metrics';
import { ChartComponents } from '@/mixins/useChartComponents';
import { PeriodType, PeriodOptions, PeriodTypes } from '@/types/chart.type';
import { getPeriodOffset } from '@/contracts/chart';

const props = defineProps({
  id: { type: String, required: true },
});


const selectePeriod = ref<PeriodType>(PeriodTypes.Today);

watch(
  () => props.id,
  () => {

    dateSelected(selectePeriod.value)
  }, { immediate: true }
)

function dateSelected(value: PeriodType) {

  console.log('dateSelected', value);
  const { from, to } = getPeriodOffset(value);

  console.log(from, ' ----', to);
  var request: DeviceMetricsRequest = {
    id: props.id,
    from: toUnix(from),
    to: toUnix(to),
  };

  store.dispatch('metrics/query', request);
}

const groupedMetrics = computed(() => {
  const results = store.getters['metrics/view'](
    props.id,
  ) as DeviceMetrics;

  return groupMetricsByType(results) as KeyValuePair<
    DeviceExposeMetrics[]
  >;
});
</script>

<template>
  <div class="col-sm-3">
    <Selection label="Select time offset:" :value="selectePeriod" @updated="dateSelected" :items="PeriodOptions">
    </Selection>
  </div>

  <div v-for="(metrics, chartType) in groupedMetrics">
    <component :is="ChartComponents[chartType]" v-bind="{ chartData: metrics }">
    </component>
  </div>
</template>
