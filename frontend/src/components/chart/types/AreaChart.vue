<script setup lang="ts">
  import { computed } from 'vue';
  import type { PropType } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import type { DeviceExposeNumericMetrics } from '@/types/metrics.type';
  import { getExposeColor, resolveChartOptions } from '@/contracts/chart';
  import { ChartTypes } from '@/types/chart.type';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeNumericMetrics>,
      required: true,
    },
  });

  const series = computed<ApexAxisChartSeries>(() => {
    if (!props.chartData?.data) return [];

    return [
      {
        name: props.chartData.name,
        data: props.chartData.data.map((point) => ({
          x: point.x,
          y: point.y,
        })),
      },
    ];
  });

  const options = computed(() => {
    const color = getExposeColor(props.chartData?.name ?? '');

    return resolveChartOptions(ChartTypes.AreaChart, {
      stroke: { colors: [color] },
      colors: [color],
    });
  });
</script>

<template>
  <BaseChart v-if="series.length > 0" :height="240" :width="'100%'" :options="options" :data="series" />
</template>
