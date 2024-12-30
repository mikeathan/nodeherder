<script setup lang="ts">
  import { PropType } from 'vue';
  import {
    DeviceExposeMetrics,
    DeviceExposeNumericMetrics,
  } from '@/types/metrics.type';
  import AreaChart from '../chart/types/AreaChart.vue';

  const props = defineProps({
    chartData: {
      type: Object as PropType<
        DeviceExposeNumericMetrics[]
      >,
      required: true,
    },
  });
</script>

<template>
  <div
    v-for="(data, chartType) in props.chartData"
    :key="chartType">
    <AreaChart :chartData="data"></AreaChart>
  </div>
</template>
