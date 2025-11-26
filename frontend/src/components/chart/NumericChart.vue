<script setup lang="ts">
  import { PropType } from 'vue';
  import { DeviceExposeNumericMetrics } from '@/types/metrics.type';
  import AreaChart from '../chart/types/AreaChart.vue';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeNumericMetrics[]>,
      required: true,
    },
  });
</script>

<template>
  <div v-for="(data, chartType) in props.chartData" :key="chartType">
    <DateRangeDisplay v-if="props.chartData.length > 0" :from="props.chartData[0].from" :to="props.chartData[0].to" />
    <AreaChart :chartData="data"></AreaChart>
  </div>
</template>
