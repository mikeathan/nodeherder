<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import VueApexCharts from 'vue3-apexcharts';

const presenceData = ref([
  // Replace with your actual presence data
  { timestamp: '2024-07-17T09:00:00', value: 1 }, // On at 9:00
  { timestamp: '2024-07-17T12:00:00', value: 0 }, // Off at 12:00
  { timestamp: '2024-07-17T14:00:00', value: 1 }, // On at 14:00
  { timestamp: '2024-07-17T18:00:00', value: 0 }, // Off at 18:00
]);

const series = [
  {
    name: 'George Washington',
    data: [
      {
        x: 'President',
        y: [
          new Date(1789, 3, 30).getTime(),
          new Date(1797, 2, 4).getTime(),
        ],
      },
    ],
  },
  // John Adams
  {
    name: 'John Adams',
    data: [
      {
        x: 'President',
        y: [
          new Date(1797, 2, 4).getTime(),
          new Date(1801, 2, 4).getTime(),
        ],
      },
      {
        x: 'Vice President',
        y: [
          new Date(1789, 3, 21).getTime(),
          new Date(1797, 2, 4).getTime(),
        ],
      },
    ],
  },
  // Thomas Jefferson
  {
    name: 'Thomas Jefferson',
    data: [
      {
        x: 'President',
        y: [
          new Date(1801, 2, 4).getTime(),
          new Date(1809, 2, 4).getTime(),
        ],
      },
      {
        x: 'Vice President',
        y: [
          new Date(1797, 2, 4).getTime(),
          new Date(1801, 2, 4).getTime(),
        ],
      },
      {
        x: 'Secretary of State',
        y: [
          new Date(1790, 2, 22).getTime(),
          new Date(1793, 11, 31).getTime(),
        ],
      },
    ],
  },
];
const chartOptions = {
  chart: {
    height: 350,
    type: 'rangeBar',
  },
  plotOptions: {
    bar: {
      horizontal: true,
      barHeight: '50%',
      rangeBarGroupRows: true,
    },
  },
  colors: [
    '#008FFB',
    '#00E396',
    '#FEB019',
    '#FF4560',
    '#775DD0',
    '#3F51B5',
    '#546E7A',
    '#D4526E',
    '#8D5B4C',
    '#F86624',
    '#D7263D',
    '#1B998B',
    '#2E294E',
    '#F46036',
    '#E2C044',
  ],
  fill: {
    type: 'solid',
  },
  xaxis: {
    type: 'datetime',
  },
  legend: {
    position: 'right',
  },
};

onBeforeMount(() => {});
</script>

<template>
  <div class="chart-container">
    <VueApexCharts
      width="800"
      height="400"
      :options="chartOptions"
      :series="series">
    </VueApexCharts>
  </div>
</template>

<!-- 
// Bar
// Line -->
