<script setup lang="ts">
  import { computed } from 'vue';
  import type { PropType } from 'vue';
  import { type MetricsType } from '@/types/metrics.type';
  import type { DeviceExposeNumericMetrics, DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import { resolveMiniChartComponentKey } from '@/utils/chart.utils';
  import { MiniChartComponents } from '@/mixins/useChartComponents';

  const props = defineProps({
    type: {
      type: String as PropType<MetricsType>,
      required: true,
    },
    data: {
      type: Object as PropType<DeviceExposeNumericMetrics | DeviceExposeBinaryMetrics>,
      required: true,
    },
    exposeName: {
      type: String,
      required: true,
    },
    unit: {
      type: String,
      default: '',
    },
    height: {
      type: Number,
      default: 150,
    },
  });

  const chartKey = computed(() => resolveMiniChartComponentKey(props.type, props.exposeName, props.unit));
  const chartComponent = computed(() => (chartKey.value ? MiniChartComponents[chartKey.value] : null));
</script>

<template>
  <component
    :is="chartComponent"
    :data="data as any"
    :exposeName="exposeName"
    :unit="unit"
    :height="height"
  />
</template>
