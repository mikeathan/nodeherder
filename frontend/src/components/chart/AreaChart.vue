<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { PropType, Ref } from 'vue';
import BaseChart from './BaseChart.vue';
import { AreaChartEntry } from '@/types/chart.type';
import { DeviceExposeMetrics } from '@/types/metrics.type';

const props = defineProps({
  chartData: {
    type: Object as PropType<DeviceExposeMetrics[]>,
    default: null,
  },
});
const chartData = ref<AreaChartEntry[]>([]);

const NEWTemperatureData = {
  name: 'Temperature',
  timestamp: [
    '2024-07-17T09:00:00',
    '2024-07-17T12:00:00',
    '2024-07-17T14:00:00',
    '2024-07-17T18:00:00',
    '2024-07-17T20:00:00',
    '2024-07-17T20:05:00',
    '2024-07-17T20:10:00',
    '2024-07-17T20:24:00',
  ],
  value: [15.1, 15.9, 16.2, 16.5, 17.1, 15.1, 15.0, 14.5],
};

watch(
  () => props.chartData,
  () => {
    if (props.chartData !== null) {
      chartData.value = transformedChartData(props.chartData);
    }
  }, { immediate: true }
)

function transformedChartData(chartData: DeviceExposeMetrics[]): AreaChartEntry[] {

  chartData.map((item) => {
    return [{
      name: item.name,
      data: item.timestamp.map((timestamp, index) => ({
        x: timestamp,
        y: item.values[index],
      })),
    }] as AreaChartEntry[];
  });
};

const chartOptions = {
  chart: {
    type: 'area',
    background: '#fff',
    toolbar: {
      show: false,
    },
  },
  dataLabels: {
    enabled: false,
  },

  stroke: {
    curve: 'smooth',
  },
  xaxis: {
    type: 'datetime',
    labels: {
      datetimeFormatter: {
        year: 'yyyy',
        month: "MMM 'yy",
        day: 'dd MMM',
        hour: 'HH:mm',
      },
    },
  },
  Tooltip: {
    x: {
      format: 'dd/MMM/yy HH:mm:ss ',
    },
  },
  responsive: [
    {
      breakpoint: undefined,
      options: {
        chart: {
          width: '100%',
        },
      },
    },
  ],
};
</script>

<template>
  chartData:{{ chartData }}
  <div class="area-chart">
    <!-- <BaseChart :data="chartData" :options="chartOptions" /> -->
  </div>
</template>
