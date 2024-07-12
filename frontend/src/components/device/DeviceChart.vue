<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import { Bar, Line, Scatter, Bubble } from 'vue-chartjs';
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
const chartOptions2 = {
  scales: {
    x: {
      type: 'time',
      time: {
        unit: 'hour',
        displayFormats: {
          hour: 'HH:mm',
        },
      },
      title: {
        display: true,
        text: 'Time',
      },
    },
    y: {
      title: {
        display: true,
        text: 'Temperature (°C)', // Adjust units as needed
      },
    },
  },
};
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
  <Line :options="chartOptions" :data="props.chartData" />
</template>

<!-- 
// Bar
// Line -->
