<script setup lang="ts">
import { ref, computed, watch, toRaw } from 'vue';
import type { PropType, Ref } from 'vue';
import BaseChart from '../BaseChart.vue';
import { TimelineChartEntry } from '@/types/chart.type';
import { DeviceExposeMetrics } from '@/types/metrics.type';

const props = defineProps({
  chartData: {
    type: Object as PropType<DeviceExposeMetrics[]>,
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

//TEMP
function addOneMinute(date: Date) {
  const newDate = new Date(date);

  newDate.setTime(newDate.getTime() + 60000);

  return newDate;
}

function transformedChartData(
  exposeMetrics: DeviceExposeMetrics[],
): TimelineChartEntry[] {
  const transformedData: TimelineChartEntry[] = [];

  exposeMetrics.forEach((item) => {
    item.timestamp.forEach((timestamp, index) => {
      const currentValue = item.values[index];

      const currentTimestamp = timestamp;
      const nextTimestamp =
        index + 1 >= item.timestamp.length
          ? addOneMinute(
              new Date(item.timestamp[index + 1]),
            ).getTime() /* TEMPORARY */
          : new Date(item.timestamp[index + 1]).getTime();

      const entry = {
        name: currentValue === 0 ? 'Off' : 'On',
        data: [
          {
            x: item.name,
            y: [
              new Date(currentTimestamp).getTime(),
              new Date(nextTimestamp).getTime(),
            ],
          },
        ],
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
    <BaseChart
      height="200"
      :data="toRaw(timelineData)"
      :options="chartOptions" />
  </div>
</template>
