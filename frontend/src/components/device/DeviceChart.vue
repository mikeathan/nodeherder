<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import { Bar, Line } from 'vue-chartjs';
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Chart,
  ChartData,
  type ChartOptions,
  ChartDataset,
} from 'chart.js';

const data2 = {
  datasets: [
    {
      label: 'brightness',
      data: [
        { x: '2024-07-04T15:00:00Z', y: 110.8 },
        { x: '2024-07-04T16:00:00Z', y: 110.3 },
        { x: '2024-07-04T17:00:00Z', y: 110.8 },
        { x: '2024-07-04T18:00:00Z', y: 110.1 },
        { x: '2024-07-04T19:00:00Z', y: 110.1 },
        { x: '2024-07-04T20:00:00Z', y: 110.7 },
      ],
    },
    {
      label: 'color_temp',
      data: [
        { x: '2024-07-04T15:00:00Z', y: 100 },
        { x: '2024-07-04T16:00:00Z', y: 150 },
        { x: '2024-07-04T17:00:00Z', y: 250 },
        { x: '2024-07-04T18:00:00Z', y: 200 },
        { x: '2024-07-04T19:00:00Z', y: 350 },
        { x: '2024-07-04T20:00:00Z', y: 100 },
      ],
    },
  ],
};

const data: ChartData = {
  labels: [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December',
  ],
  datasets: [
    {
      label: 'Data One',
      backgroundColor: '#f87979',
      data: [
        40, 20, 12, 39, 10, 40, 39, 80, 40, 20, 12, 11,
      ],
    },
    {
      label: 'Data TWO',
      backgroundColor: 'white',
      data: [12, 17, 19, 30, 35, 37, 29, 59],
    },
  ],
};

const props = defineProps({
  chartData: {
    type: Object as PropType<any>,
    default: null,
  },
  options: {
    type: Object,
    default: null,
  },
});

const chartOptions = ref({
  responsive: true,
  maintainAspectRatio: true,
  plugins: {
    legend: {
      display: true,
    },
  },
});

const chartData = ref({
  labels: ['January', 'February', 'March'],
  datasets: [{ data: [50, 30, 35] }],
}) as Ref<ChartData>;

onBeforeMount(() => {
  ChartJS.register(
    Title,
    Tooltip,
    Legend,
    BarElement,
    CategoryScale,
    PointElement,
    LineElement,
    LinearScale,
  );
});
</script>

<template>
  <Line :options="chartOptions" :data="data" />
</template>
