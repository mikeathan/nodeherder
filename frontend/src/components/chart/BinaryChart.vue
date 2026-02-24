<script setup lang="ts">
  import { PropType } from 'vue';
  import type { DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import DynamicBinaryChart from './DynamicBinaryChart.vue';

  defineProps({
    chartData: {
      type: Array as PropType<DeviceExposeBinaryMetrics[]>,
      required: true,
    },
  });
</script>

<template>
  <div v-for="chart in chartData" :key="chart.name" class="binary-chart-wrapper">
    <DynamicBinaryChart :chartData="chart" />
  </div>
</template>

<style scoped>
  .binary-chart-wrapper {
    padding: 8px 0 18px;
  }
</style>
