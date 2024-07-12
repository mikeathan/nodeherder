<script setup lang="ts">
import { store } from '../../store/index';
import { computed, onMounted, ref } from 'vue';
import {
  DeviceMetricsRequest,
  DeviceMetrics,
} from '@/types/metrics';
import Selection from '../input/Selection.vue';
import DeviceChart from '../device/DeviceChart.vue';
import { KeyyValuePair } from '@/types/types';
import { ChartColor } from '@/types/chart.type';
import { toUnix } from '@/utils/date.utils';
import { chartColors } from '@/contracts/chart';

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

const deviceMetrics = computed(() => {
  const results = store.getters['metrics/view'](
    props.id,
  ) as DeviceMetrics;
  if (results != null) {
    const chartData = {
      datasets: results.expose.map((expose, idx) => ({
        label: expose.name,
        backgroundColor: chartColors[idx].backgroundColor,
        borderColor: chartColors[idx].borderColor,
        data: expose.timestamp.map((timestamp, index) => ({
          x: timestamp,
          y: expose.values[index],
        })),
      })),
    };
    return chartData;
  }
  return { datasets: [] };
});
</script>
<style scoped></style>

<template>
  <div class="col-sm-3">
    <Selection
      label="Select time offset:"
      @updated="dateSelected"
      :items="historySelection">
    </Selection>
  </div>
  <DeviceChart :chartData="deviceMetrics"></DeviceChart>
</template>
