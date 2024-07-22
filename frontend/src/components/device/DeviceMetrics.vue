<script setup lang="ts">
import { store } from '../../store/index';
import { computed, onMounted, ref } from 'vue';
import {
  DeviceMetricsRequest,
  DeviceMetrics,
} from '@/types/metrics';
import Selection from '../input/Selection.vue';
import TimelineChart from '../chart/TimelineChart.vue';
import AreaChart from '../chart/AreaChart.vue';

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
const timelineDataset = computed(() => {
  const results = store.getters['metrics/view'](
    props.id,
  ) as DeviceMetrics;
  if (results != null) {
    const datasets = results.expose.map((expose, idx) => ({
      label: expose.name,

      data: expose.timestamp.map((timestamp, index) => ({
        x: timestamp,
        y: expose.values[index],
      })),
    }));

    const parsedTimestamps: Date[] = datasets.reduce(
      (acc: any, dataset) => {
        dataset.data.forEach((datapoint) => {
          const timestamp = new Date(datapoint.x); // Assuming timestamps are strings in YYYY-MM-DD format
          acc.push(timestamp);
        });
        return acc;
      },
      [],
    );

    const labels = [...new Set(parsedTimestamps)].map(
      (timestamp) => timestamp.toLocaleDateString(),
    );
    const chartData = {
      data: {
        labels,
        datasets: datasets.map((dataset, idx) => ({
          ...dataset,
          backgroundColor: chartColors[idx].backgroundColor,
          borderColor: chartColors[idx].borderColor,
          pointRadius: 0,
          stack: 'data',
        })),
      },
    };
    return chartData;
  }
  return { datasets: [] };
});

const timelineOptions = {
  scales: {
    x: {
      stacked: true, // Enable stacking for the x-axis
    },
    y: {
      beginAtZero: true, // Start y-axis at 0 for better visualization
    },
  },
};

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

  <!-- <AreaChart :chartData="deviceMetrics" :options="timelineOptions"></AreaChart> -->
  <TimelineChart
    :chartData="deviceMetrics"
    :options="timelineOptions"></TimelineChart>
</template>
