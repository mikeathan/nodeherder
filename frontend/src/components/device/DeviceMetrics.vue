<script setup lang="ts">
import { store } from '../../store/index';
import { computed, onMounted, ref } from 'vue';
import {
  DeviceMetricsRequest,
  DeviceMetrics,
  DeviceExposeMetrics,
} from '@/types/metrics.type';
import Selection from '../input/Selection.vue';
import TimelineChart from '../chart/TimelineChart.vue';
import AreaChart from '../chart/AreaChart.vue';

import { KeyyValuePair } from '@/types/types';
import { toUnix } from '@/utils/date.utils';
import { groupMetricsByType } from '@/contracts/metrics';
import { ChartComponents } from '@/mixins/useChartComponents';

const props = defineProps({
  id: { type: String, required: true },
});

const historySelection: string[] = [
  'last hour',
  'last day',
  'last week',
  'last month',
];
const timeOffsets: KeyyValuePair<number> = {
  'last hour': 1,
  'last day': 24,
  'last week': 168,
  'last month': 720,
};
const fromDate = ref<Date | null>(null);
const toDate = ref<Date | null>(null);

function dateSelected(value: any) {
  const dateOffset = timeOffsets[value];

  toDate.value = new Date();
  fromDate.value = new Date();
  fromDate.value.setHours(
    toDate.value.getHours() - dateOffset,
  );

  var request: DeviceMetricsRequest = {
    id: props.id,
    from: toUnix(fromDate.value ?? new Date()), // temp
    to: toUnix(toDate.value ?? new Date()), // temp
  };

  store.dispatch('metrics/query', request);
}

const groupedMetrics = computed(() => {
  const results = store.getters['metrics/view'](
    props.id,
  ) as DeviceMetrics;

  return groupMetricsByType(results) as KeyyValuePair<DeviceExposeMetrics[]>;
});


</script>
<style scoped></style>

<template>
  <div class="col-sm-3">
    <Selection label="Select time offset:" @updated="dateSelected" :items="historySelection">
    </Selection>
  </div>

  <div v-for="(metric, chartType)  in groupedMetrics">
    <component :is="ChartComponents[chartType]" v-bind="{ chartData: metric }">
    </component>
  </div>
  <!-- <AreaChart :chartData="deviceMetrics" ></AreaChart> -->
  <!-- <TimelineChart :chartData="deviceMetrics"></TimelineChart> -->

</template>
