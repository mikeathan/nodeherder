<script setup lang="ts">
  import { PropType } from 'vue';
  import { DeviceExposeBinaryMetrics, DeviceExposeMetrics } from '@/types/metrics.type';
  import TimelineChart from '../chart/types/TimelineChart.vue';
  import DateRangeDisplay from '../chart/DateRangeDisplay.vue';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeBinaryMetrics[]>,
      required: true,
    },
  });
</script>

<template>
  <DateRangeDisplay v-if="props.chartData.length > 0" :from="props.chartData[0].from" :to="props.chartData[0].to" />
  <TimelineChart :chartData="props.chartData"></TimelineChart>
</template>
