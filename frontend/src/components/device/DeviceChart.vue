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
  ],
};

const props = defineProps({
  chartData: {
    type: Object as PropType<ChartData>,
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
      display: false,
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
  <Line :options="chartOptions" :data="props.chartData" />
</template>
