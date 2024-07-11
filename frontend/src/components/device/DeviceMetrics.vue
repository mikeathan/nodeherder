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
import { toUnix } from '@/utils/date.utils';
import { Chart, ChartData } from 'chart.js';

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

const chartData = ref<KeyyValuePair<DeviceMetricsRequest>>(
  {} as KeyyValuePair<DeviceMetricsRequest>,
);

const deviceMetrics = computed(() => {
  const results = store.getters['metrics/view'](
    props.id,
  ) as DeviceMetrics;
  if (results != null) {
    // brightness
    // timestamp : 1
    // value: 1
    // timestamp: 2
    // value :2

    // type: 'scatter',
    // datasets: [
    //     {
    //       label: 'Data One',
    //       backgroundColor: '#f87979',
    //       data: [],
    //     },
    //   ],

    const datasets = results.expose.map((expose) => ({
      label: expose.name,
      data: expose.timestamp.map((timestamp, index) => ({
        x: timestamp,
        y: expose.values[index],
      })),
    }));
    const chartData = {
      datasets: datasets,
    };
    return chartData;
    // const expose = results.expose[0];
    // const chartData = {
    //   labels: expose.timestamp, // THIS IS WRONG HERE
    //   datasets: results.expose.map((expose) => ({
    //     label: expose.name,
    //     data: expose.timestamp.map((timestamp, index) => ({
    //       x: timestamp,
    //       y: expose.values[index],
    //     })),
    //     borderColor: 'rgba(75, 192, 192, 0.8)', // Example blue color
    //   })),
    // };
    // results.expose.forEach((expose) => {
    //   expose.timestamp.forEach((timestamp) => {});
    // });
  }
  //     var request: DeviceMetricsRequest = {
  //         id: props.id,
  //         from: toUnix(fromDate.value ?? new Date()), // temp
  //         to: toUnix(toDate.value ?? new Date()), // temp
  //     };

  //     store.dispatch('metrics/query', request);
  // }

  return null;
  // return results;
});

// const data: ChartData = {
//   labels: [
//     'January',
//     'February',
//     'March',
//     'April',
//     'May',
//     'June',
//     'July',
//     'August',
//     'September',
//     'October',
//     'November',
//     'December',
//   ],
//   datasets: [
//     {
//       label: 'Data One',
//       backgroundColor: '#f87979',
//       data: [
//         40, 20, 12, 39, 10, 40, 39, 80, 40, 20, 12, 11,
//       ],
//     },
//   ],
// };
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
  {{ deviceMetrics }}

  <DeviceChart :chartData="{ deviceMetrics }"></DeviceChart>
</template>
