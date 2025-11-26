<script setup lang="ts">
  import { computed } from 'vue';
  import type { PropType } from 'vue';
  import MiniNumericChart from './MiniNumericChart.vue';
  import MiniBinaryChart from './MiniBinaryChart.vue';
  import { MetricsTypes, type MetricsType } from '@/types/metrics.type';
  import type { DeviceExposeNumericMetrics, DeviceExposeBinaryMetrics } from '@/types/metrics.type';

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

  const isNumeric = computed(() => props.type === MetricsTypes.Numeric);
  const isBinary = computed(() => props.type === MetricsTypes.Binary);
</script>

<template>
  <MiniNumericChart v-if="isNumeric" :data="data as any" :exposeName="exposeName" :unit="unit" :height="height" />
  <MiniBinaryChart v-else-if="isBinary" :data="data as any" :exposeName="exposeName" :height="height" />
</template>
