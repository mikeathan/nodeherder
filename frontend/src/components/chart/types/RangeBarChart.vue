<script setup lang="ts">
  import { computed } from 'vue';
  import type { PropType } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import type { DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import { getExposeBinaryColour, resolveChartOptions } from '@/contracts/chart';
  import { ChartTypes } from '@/types/chart.type';
  import { normalizeBinaryEvents, mergeBinaryFlickers, toBinaryRangeBarData } from '@/utils/chart.utils';
  import { formatDuration, formatTime, parseTimestamp } from '@/utils/date.utils';

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
        custom: ({ seriesIndex, dataPointIndex, w }: any) => {
          const data = w.globals.initialSeries?.[seriesIndex]?.data?.[dataPointIndex];
          if (!data) return '';

          const [start, end] = data.y;
          const durationStr = formatDuration(end - start);
          const startTime = formatTime(start);
          const endTime = formatTime(end);

          return `
          <div style="padding:8px;font-size:12px;background:#1e1e1e;color:#fff;border-radius:4px;">
            <div style="margin-bottom:4px;"><strong>${data.x}</strong></div>
            <div>${startTime} → ${endTime}</div>
            <div style="color:#aaa;font-size:11px;">Duration: ${durationStr}</div>
          </div>
        `;
        },
      },
    });
  });
</script>

<template>
  <BaseChart v-if="series.length > 0" :options="options" :data="series" width="100%" height="250" />
</template>
