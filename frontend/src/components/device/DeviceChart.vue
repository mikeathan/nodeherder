<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import { Bar, Line, Scatter,} from 'vue-chartjs';

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
  TimeScale,
  Colors,
  
  
} from 'chart.js';
import moment from 'moment';
import 'chartjs-adapter-moment';
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



const timelineOptions = {
  responsive: true,
  maintainAspectRatio: true,
  scales: {
    x: {
      type: 'time', // Set the x-axis type to 'time' for timestamps
      time: {
        unit: 'day', // Adjust unit based on your timestamps (e.g., 'month', 'hour')
      },
    },
    y: {
      stacked: true, // Enable stacking for bars
    },
  },
}
const chartOptions = ref({
  responsive: true,
  maintainAspectRatio: true,
  scales: {
    x: {
      type: 'time',
      time: {
        unit: 'hour', // same here
        displayFormats: {
          hour: 'HH:mm', // pass fomat here from props
        },
        title: {
          display: true,
          text: 'Time',
        },
      },
    },
    y: {
      title: {
        display: true,
      },
    },
  },
  plugins: {
    legend: {
      display: true,
      //usePointStyle: true,
    },
  },
});

onBeforeMount(() => {
  ChartJS.register(
    Colors,
    Title,
    Tooltip,
    Legend,
    BarElement,
    CategoryScale,
    PointElement,
    LineElement,
    LinearScale,
    BarElement,
    TimeScale,
  );
});
</script>

<template>
  <Bar :options="timelineOptions" :data="props.chartData" />
</template>

<!-- 
// Bar
// Line -->
