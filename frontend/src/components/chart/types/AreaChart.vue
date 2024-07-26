<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { PropType, Ref } from 'vue';
import BaseChart from '../BaseChart.vue';
import { AreaChartEntry } from '@/types/chart.type';
import { DeviceExposeMetrics } from '@/types/metrics.type';

const props = defineProps({
  chartData: {
    type: Object as PropType<DeviceExposeMetrics>,
    default: null,
  },
});

const chartData = ref<AreaChartEntry[]>([]);

watch(
  () => props.chartData,
  () => {
    if (props.chartData !== null) {
      chartData.value = transformedChartData(
        props.chartData,
      );
    }
  },
  { immediate: true },
);

function transformedChartData(
  chartData: DeviceExposeMetrics,
): AreaChartEntry[] {
  return [
    {
      name: chartData.name,
      data: chartData.timestamp.map((timestamp, index) => ({
        x: timestamp,
        y: chartData.values[index],
      })),
    },
  ] as AreaChartEntry[];
}

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
  legend: {
    position: 'top',
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
  <div class="area-chart">
    <BaseChart
      height="200"
      :data="chartData"
      :options="chartOptions" />
  </div>
</template>
