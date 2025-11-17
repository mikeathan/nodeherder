<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import BaseChart from '../BaseChart.vue';
import { getExposeColor } from '@/contracts/chart';

const props = defineProps({
  chartData: {
    type: Object,
    required: true,
  },
});

const series = ref([] as ApexAxisChartSeries);

watch(
  () => props.chartData,
  () => {
    if (props.chartData) {
      series.value = [
        {
          name: props.chartData.name,
          data: props.chartData.data.map((p: { x: number; y: number }) => ({
            x: p.x,
            y: p.y,
          })),
        },
      ];
    }
  },
  { immediate: true }
);

// Color for both stroke + gradient
const color = computed(() => getExposeColor(props.chartData.name));

const options = computed(() => ({
  chart: {
    type: 'area',
    background: 'transparent',
    toolbar: { show: false },
    zoom: { enabled: false },
  },

  stroke: {
    curve: 'smooth',
    width: 1.5,
    colors: [color.value],
  },

  fill: {
    type: 'gradient',
    gradient: {
      shadeIntensity: 0.15,
      opacityFrom: 0.15,
      opacityTo: 0,
      stops: [0, 100],
    },
  },

  markers: {
    size: 0,
    hover: {
      size: 4,
    },
  },

  dataLabels: { enabled: false },

  grid: {
    borderColor: 'rgba(255,255,255,0.08)',
    strokeDashArray: 3,
  },

  xaxis: {
    type: 'datetime',
    labels: {
      datetimeUTC: false,
      style: { fontSize: '10px', colors: '#aaa' },
    },
  },

  yaxis: {
    decimalsInFloat: 0,
    labels: {
      style: { fontSize: '10px', colors: '#aaa' },
    },
  },

  tooltip: {
    theme: 'dark',
    shared: false,
    marker: { show: false },
    y: {
      formatter: (v: number | null) => (v === null ? '' : v.toFixed(1)),
    },
    x: {
      format: 'dd MMM HH:mm',
    },
  },

  legend: {
    show: true,
    position: 'top',
    labels: { colors: '#ccc' },
  },
}));

</script>

<template>
  <BaseChart height="240" :options="options" :data="series" />
</template>
