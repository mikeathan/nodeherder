<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { BinaryDataPoint } from '@/types/metrics.type';
  import { getExposeBinaryColour } from '@/contracts/chart';

  const props = defineProps({
    data: {
      type: Array as PropType<BinaryDataPoint[]>,
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
  });

  const color = computed(() => {
    const colors = getExposeBinaryColour(props.exposeName);
    return colors.on;
  });

  // Transform binary data into step line (Home Assistant style)
  const chartData = computed(() => {
    if (!props.data || props.data.length === 0) return [];

    const points: Array<{ x: number; y: number }> = [];

    props.data.forEach((point) => {
      // Convert binary to 0/1
      const value =
        point.value.toLowerCase() === 'true' || point.value.toLowerCase() === 'on' || point.value === '1' ? 1 : 0;

      points.push({
        x: point.timestamp,
        y: value,
      });
    });

    // Add current time with last value
    if (points.length > 0) {
      points.push({
        x: Date.now(),
        y: points[points.length - 1].y,
      });
    }

    return [
      {
        name: props.exposeName,
        data: points,
      },
    ];
  });

  const chartOptions = computed(() => ({
    chart: {
      type: 'line',
      sparkline: {
        enabled: true,
      },
      animations: {
        enabled: false,
      },
    },
    stroke: {
      curve: 'stepline',
      width: 2,
    },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.7,
        opacityTo: 0.15,
        stops: [0, 100],
      },
    },
    colors: [color.value],
    tooltip: {
      enabled: false,
    },
    xaxis: {
      type: 'datetime',
    },
    yaxis: {
      min: 0,
      max: 1,
    },
    markers: {
      size: 0,
    },
  }));
</script>

<template>
  <div class="binary-sparkline-chart">
    <VueApexCharts :height="height" :options="chartOptions" :series="chartData" />
  </div>
</template>

<style scoped>
  .binary-sparkline-chart {
    width: 100%;
    height: 100%;
    position: absolute;
    bottom: 0;
    left: 0;
    opacity: 0.5;
    pointer-events: none;
  }
</style>
