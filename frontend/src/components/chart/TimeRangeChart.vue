<script setup lang="ts">
import { PropType } from 'vue';
import { DeviceExposeBinaryMetrics, DeviceExposeMetrics } from '@/types/metrics.type';
import TimelineChart from '../chart/types/TimelineChart.vue';

const props = defineProps({
  chartData: {
    type: Object as PropType<DeviceExposeBinaryMetrics[]>,
    required: true,
  },
});
</script>

<template>
  <TimelineChart :chartData="props.chartData"></TimelineChart>
</template>
