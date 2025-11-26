<script setup lang="ts">
  import { PropType } from 'vue';
  import { DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import RangeBarChart from '../chart/types/RangeBarChart.vue';
  import DateRangeDisplay from '../chart/DateRangeDisplay.vue';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeBinaryMetrics[]>,
      required: true,
    },
  });
</script>

<template>
  <div v-for="(data, chartType) in props.chartData" :key="chartType">
    <DateRangeDisplay v-if="props.chartData.length > 0" :from="props.chartData[0].from" :to="props.chartData[0].to" />
    <RangeBarChart :chartData="data"></RangeBarChart>
  </div>
</template>
