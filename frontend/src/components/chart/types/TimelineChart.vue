<script setup lang="ts">
  import { ref, watch, toRaw, computed } from 'vue';
  import type { PropType, Ref } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import { ChartTypes, TimelineChartEntry } from '@/types/chart.type';
  import { DeviceExposeBinaryMetrics, DeviceExposeMetrics } from '@/types/metrics.type';
  import { ColorValue } from '@/types/color.type';
  import { getExposeBinaryColour, resolveChartOptions } from '@/contracts/chart';
  import { ExposeTypes } from '@/types/device.type';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeBinaryMetrics[]>,
      default: null,
    },
  });

  const timelineData = ref<TimelineChartEntry[]>([]);

  watch(
    () => props.chartData,
    () => {
      if (props.chartData !== null) {
        timelineData.value = transformedChartData(props.chartData);
      }
    },
    { immediate: true }
  );

  function transformedChartData(exposeMetrics: DeviceExposeBinaryMetrics[]): TimelineChartEntry[] {
    return exposeMetrics.map((item) => ({
      name: item.name,
      data: item.data.map((point) => {
        const isOn = point.x === 'true';
        const color = getExposeBinaryColour(item.name);
        return {
          x: item.name,
          y: point.y,
          state: point.x,
          fillColor: isOn ? color.on : color.off,
        };
      }),
    }));
  }
  const timelineTooltip = {
    custom: ({ seriesIndex, dataPointIndex, w }: any) => {
      const data = w.globals.initialSeries[seriesIndex].data[dataPointIndex];
      if (!data || !Array.isArray(data.y)) return '';
      const [startMs, endMs] = data.y;
      const start = new Date(startMs);
      const end = new Date(endMs);
      const durationMs = endMs - startMs;
      const mins = Math.floor(durationMs / 60000);
      const hrs = Math.floor(mins / 60);
      const rem = mins % 60;
      const duration = hrs > 0 ? `${hrs}h ${rem}m` : `${mins}m`;
      const fmt = (d: Date) => d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
      const label = data.x ?? '';
      const state = data.state === 'true' ? 'On' : data.state === 'false' ? 'Off' : data.state ?? '';

      return `
      <div style="padding:6px;font-size:12px;background:#1e1e1e;color:#fff;border-radius:4px">
        <b>${label}</b>: ${state}<br/>
        ${fmt(start)} → ${fmt(end)}<br/>
        <small>${duration}</small>
      </div>
    `;
    },
  };

  const chartOptions = resolveChartOptions(ChartTypes.TimelineChart, { tooltip: timelineTooltip });
</script>

<template>
  <div style="width: 100%; height: 240px;">
  <BaseChart
    width="100%"
    height="100%"
    :data="toRaw(timelineData)"
    :options="chartOptions"
  />
</div>
</template>
