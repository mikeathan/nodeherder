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
} from 'chart.js';

const props = defineProps({
  chartData: {
    type: Object as PropType<any>, // ChartData<'bar'
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
    BarElement,
  );
});
</script>

<template>
  <Line :options="chartOptions" :data="props.chartData" />
</template>

<!-- 
// Bar
// Line -->