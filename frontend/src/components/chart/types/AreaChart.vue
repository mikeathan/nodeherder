<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { PropType, Ref } from 'vue';
import BaseChart from '../BaseChart.vue';
import { AreaChartEntry } from '@/types/chart.type';
import { DeviceExposeMetrics, DeviceExposeNumericMetrics } from '@/types/metrics.type';

const props = defineProps({
  chartData: {
    type: Object as PropType<DeviceExposeNumericMetrics>,
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
  chartData: DeviceExposeNumericMetrics,
): AreaChartEntry[] {
  return [
    {
      name: chartData.name,
      data: chartData.data.map((point) => ({
        x: point.x,
        y: point.y,
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
  legend: {
    position: 'top'
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
  yaxis: {
    title: {
      text: props.chartData.name,
    }
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
  <div class="area-chart">
    <BaseChart height="200" :data="chartData" :options="chartOptions" />
  </div>
</template>
