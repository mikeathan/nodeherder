<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { NumericDataPoint } from '@/types/metrics.type';
  import { getExposeColor } from '@/contracts/chart';

  const props = defineProps({
    data: {
      type: Array as PropType<NumericDataPoint[]>,
      required: true,
    },
    exposeName: {
      type: String,
      required: true,
    },
    height: {
      type: Number,
      default: 60,
    },
    color: {
      type: String,
      default: undefined,
    },
  });

  // Downsample data for smoother, cleaner sparkline (Home Assistant style)
  const downsampledData = computed(() => {
    if (!props.data || props.data.length === 0) return [];

    // Target ~20-30 points for smooth curve
    const targetPoints = 25;
    const step = Math.max(1, Math.floor(props.data.length / targetPoints));

    const sampled: NumericDataPoint[] = [];
    for (let i = 0; i < props.data.length; i += step) {
      sampled.push(props.data[i]);
    }

    // Always include last point
    if (sampled[sampled.length - 1] !== props.data[props.data.length - 1]) {
      sampled.push(props.data[props.data.length - 1]);
    }

    return sampled;
  });

  const chartData = computed(() => {
    return [
      {
        name: props.exposeName,
        data: downsampledData.value.map((point) => ({
          x: point.x,
          y: point.y,
        })),
      },
    ];
  });

  const lineColor = computed(() => {
    if (props.color) return props.color;
    const color = getExposeColor(props.exposeName);
    return color;
  });

  const chartOptions = computed(() => ({
    chart: {
      type: 'area',
      sparkline: {
        enabled: true,
      },
      animations: {
        enabled: false,
      },
    },
    stroke: {
      curve: 'smooth',
      width: 2,
    },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.7,
        opacityTo: 0.1,
        stops: [0, 100],
      },
    },
    colors: [lineColor.value],
    tooltip: {
      enabled: false,
    },
    xaxis: {
      type: 'datetime',
    },
    markers: {
      size: 0,
    },
  }));
</script>

<template>
  <div class="sparkline-chart">
    <VueApexCharts :height="height" :options="chartOptions" :series="chartData" />
  </div>
</template>

<style scoped>
  .sparkline-chart {
    width: 100%;
    height: 100%;
    position: absolute;
    bottom: 0;
    left: 0;
    opacity: 0.6;
    pointer-events: none;
  }
</style>
