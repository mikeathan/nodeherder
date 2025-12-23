<script setup lang="ts">
  import { computed } from 'vue';
  import type { PropType } from 'vue';
  import MiniNumericChart from './MiniNumericChart.vue';
  import MiniRealtimeChart from './MiniRealtimeChart.vue';
  import MiniBinaryChart from './MiniBinaryChart.vue';
  import MiniEnergyChart from './MiniEnergyChart.vue';
  import { MetricsTypes, type MetricsType } from '@/types/metrics.type';
  import type { DeviceExposeNumericMetrics, DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import { isEnergyExpose } from '@/utils/chart.utils';

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
  const isEnergy = computed(() => isNumeric.value && isEnergyExpose(props.exposeName, props.unit));
  const isInstant = computed(() => {
    const unit = props.unit.toLowerCase();
    if (unit === 'a' || unit === 'ma' || unit === 'v' || unit === 'kv') return true;
    const name = props.exposeName.toLowerCase();
    return name.includes('current') || name.includes('voltage');
  });
</script>

<template>
  <MiniEnergyChart v-if="isEnergy" :data="data as any" :exposeName="exposeName" :unit="unit" :height="height" />
  <MiniRealtimeChart
    v-else-if="isNumeric && isInstant"
    :data="data as any"
    :exposeName="exposeName"
    :unit="unit"
    :height="height"
  />
  <MiniNumericChart v-else-if="isNumeric" :data="data as any" :exposeName="exposeName" :unit="unit" :height="height" />
  <MiniBinaryChart v-else-if="isBinary" :data="data as any" :exposeName="exposeName" :height="height" />
</template>
