<script setup lang="ts">
import { PropType } from 'vue';
import { DeviceExposeMetrics } from '@/types/metrics.type';
import TimelineChart from '../chart/types/TimelineChart.vue';

const props = defineProps({
  chartData: {
    type: Object as PropType<DeviceExposeMetrics[]>,
    required: true,
  },
});
</script>

<template>
  <TimelineChart
    :chartData="props.chartData"></TimelineChart>
</template>
