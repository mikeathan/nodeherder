<script setup lang="ts">
import { ref, watch, toRaw } from 'vue';
import type { PropType, Ref } from 'vue';
import BaseChart from '../BaseChart.vue';
import { TimelineChartEntry } from '@/types/chart.type';
import { DeviceExposeBinaryMetrics, DeviceExposeMetrics } from '@/types/metrics.type';

const props = defineProps({
  chartData: {
    type: Object as PropType<DeviceExposeBinaryMetrics[]>,
    default: null,
  },
});

const timelineData = ref<TimelineChartEntry[]>([]);

watch(
  () => props.chartData,
  () => {
    if (props.chartData !== null) {
      timelineData.value = transformedChartData(
        props.chartData,
      );
    }
  },
  { immediate: true },
);

function transformedChartData(
  exposeMetrics: DeviceExposeBinaryMetrics[],
): TimelineChartEntry[] {
  const transformedData: TimelineChartEntry[] = [];

  exposeMetrics.forEach((item) => {
    item.data.forEach((point) => {
      const entry = {
        name: point.x,
        data: [
          {
            x: item.name,
            y: point.y
          },
        ]
      };

      transformedData.push(entry);
    });
  });

  return transformedData;
}

const chartOptions = {
  chart: {
    type: 'rangeBar',
    background: '#fff',
    toolbar: {
      show: false,
    },
  },
  tooltip: {
    x: {
      format: 'dd/MMM/yy HH:mm:ss ',
    },
  },
  plotOptions: {
    bar: {
      horizontal: true,
      barHeight: '50%',
      rangeBarGroupRows: true,
    },
    fill: {
      type: 'solid',
    },
    xaxis: {
      type: 'datetime',
    },
  },
  colors: ['#FF4560', '#00E396'],
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
  stroke: {
    width: 1,
  },
  fill: {
    type: 'solid',
    opacity: 0.6,
  },
  legend: {
    show: false,
  },
  responsive: [
    {
      breakpoint: undefined, // Matches all screens
      options: {
        chart: {
          width: '100%', // Set chart width to 100% for all screens
        },
      },
    },
  ],
};
</script>

<template>
  <div class="timeline-chart">
    <BaseChart height="200" :data="toRaw(timelineData)" :options="chartOptions" />
  </div>
</template>
