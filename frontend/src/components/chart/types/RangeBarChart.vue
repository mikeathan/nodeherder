<script setup lang="ts">
  import { computed } from 'vue';
  import type { PropType } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import type { DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import { getExposeBinaryColour, resolveChartOptions } from '@/contracts/chart';
  import { ChartTypes } from '@/types/chart.type';
  import {
    normalizeBinaryEvents,
    mergeBinaryFlickers,
    toBinaryRangeBarData,
    renderRangeTooltip,
  } from '@/utils/chart.utils';
  import { parseTimestamp } from '@/utils/date.utils';

  const FLICKER_THRESHOLD_MS = 15_000;

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeBinaryMetrics>,
      required: true,
    },
  });

  const series = computed<ApexAxisChartSeries>(() => {
    if (!props.chartData?.data) return [];

    const { data, from, to, name } = props.chartData;
    const now = Date.now();
    const fromTimestamp = parseTimestamp(from, now - 86400000);
    const toTimestamp = parseTimestamp(to, now);

    const colors = getExposeBinaryColour(name ?? '');
    const normalizedRanges = normalizeBinaryEvents(data, fromTimestamp, toTimestamp);
    const mergedRanges = mergeBinaryFlickers(normalizedRanges, FLICKER_THRESHOLD_MS);
    const apexData = toBinaryRangeBarData(mergedRanges, colors.on, colors.off);

    return [
      {
        name: name ?? 'State',
        data: apexData,
      },
    ];
  });

  const options = computed(() => {
    return resolveChartOptions(ChartTypes.BinaryChart, {
      tooltip: {
        theme: 'dark',
        x: { format: 'dd MMM HH:mm' },
        custom: ({ w, seriesIndex, dataPointIndex }: { w: any; seriesIndex: number; dataPointIndex: number }) => {
          const d = w.config.series[seriesIndex].data[dataPointIndex];
          const start: number = Array.isArray(d.y) ? d.y[0] : d.y?.from ?? d.y ?? 0;
          const end: number = Array.isArray(d.y) ? d.y[1] : d.y?.to ?? d.y ?? 0;
          const label: string = typeof d.x === 'string' ? d.x : d.x?.toString?.() ?? '';
          return renderRangeTooltip(props.chartData.name ?? 'State', label, start, end);
        },
      },
    });
  });
</script>

<template>
  <BaseChart v-if="series.length > 0" :options="options" :data="series" width="100%" height="250" />
</template>
