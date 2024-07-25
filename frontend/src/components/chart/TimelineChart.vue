<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { PropType, Ref } from 'vue';
import BaseChart from './BaseChart.vue';
import { TimelineChartEntry } from '@/types/chart.type';
import { DeviceExposeMetrics } from '@/types/metrics.type';

const props = defineProps({
  chartData: {
    type: Object as PropType<DeviceExposeMetrics[]>,
    default: null,
  },
});

const chartData = ref<TimelineChartEntry[]>([]);

// watch(
//   () => props.chartData,
//   () => {
//     if (props.chartData !== null) {
//       chartData.value = transformedChartData(
//         props.chartData,
//       );
//     }
//   },
//   { immediate: true },
// );

const PresenceData = [
  {
    name: 'Presence Kitchen',
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
    values: [1, 0, 1, 0, 1, 0, 1, 0],
  },
  {
    name: 'Presence Living Room',
    timestamp: [
      '2024-07-17T11:00:00',
      '2024-07-17T12:00:00',
      '2024-07-17T13:00:00',
      '2024-07-17T14:00:00',
      '2024-07-17T20:00:00',
      '2024-07-17T21:05:00',
      '2024-07-17T22:10:00',
      '2024-07-17T23:24:00',
    ],
    values: [1, 0, 1, 0, 1, 0, 1, 0],
  },
];

//TEMP
function addOneMinute(date: Date) {
  const newDate = new Date(date);

  newDate.setTime(newDate.getTime() + 60000);

  return newDate;
}

const transformedChartData = computed(() => {
  const transformedData: TimelineChartEntry[] = [];

  PresenceData.forEach((item) => {

    item.timestamp.forEach((timestamp, index) => {
      const currentValue = item.values[index];

      const currentTimestamp = timestamp;
      const nextTimestamp =
        index + 1 >= item.timestamp.length
          ? addOneMinute(new Date(item.timestamp[index + 1])).getTime() /* TEMPORARY */
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
      }
      transformedData.push(entry);
    });

  });

  return transformedData;
});
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
    <BaseChart height="200" :data="transformedChartData" :options="chartOptions" />
  </div>
</template>
