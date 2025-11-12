<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import type { PropType } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import { ChartTypes, TimelineChartEntry } from '@/types/chart.type';
  import { DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import { getExposeBinaryColour, resolveChartOptions } from '@/contracts/chart';

  type BinaryPoint = DeviceExposeBinaryMetrics['data'][number];
  type RowLayout = 'single' | 'split';
  const MERGE_TOLERANCE_MS = 30 * 1000;

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeBinaryMetrics[]>,
      default: () => [],
    },
    defaultRowLayout: {
      type: String as PropType<RowLayout>,
      default: 'single',
    },
  });

  const rowLayout = ref<RowLayout>(props.defaultRowLayout);
  watch(
    () => props.defaultRowLayout,
    (layout) => {
      rowLayout.value = layout ?? 'single';
    }
  );
  const isSingleRow = computed(() => rowLayout.value === 'single');
  const layoutLabel = computed(() => (isSingleRow.value ? 'Single row' : 'Split rows'));

  function toggleRowLayout() {
    rowLayout.value = isSingleRow.value ? 'split' : 'single';
  }

  const timelineTooltip = {
    custom: ({ seriesIndex, dataPointIndex, w }: any) => {
      const data = w.globals.initialSeries[seriesIndex]?.data?.[dataPointIndex];
      if (!data || !Array.isArray(data.y)) return '';

      const [startMs, endMs] = data.y;
      const durationMs = Math.max(0, endMs - startMs);
      const mins = Math.floor(durationMs / 60000);
      const hrs = Math.floor(mins / 60);
      const rem = mins % 60;
      const duration = hrs > 0 ? `${hrs}h ${rem}m` : `${mins}m`;
      const fmt = (value: number) =>
        new Date(value).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });

      return `
        <div style="padding:6px;font-size:12px;background:#1e1e1e;color:#fff;border-radius:4px">
          <b>${data.label ?? ''}</b>: ${formatStateLabel(data.state)}<br/>
          ${fmt(startMs)} → ${fmt(endMs)}<br/>
          <small>${duration}</small>
        </div>
      `;
    },
  };

  const timelineData = computed<TimelineChartEntry[]>(() =>
    (props.chartData ?? []).map((expose) => buildSeries(expose, isSingleRow.value))
  );

  const timelineExtent = computed(() => {
    const timestamps = timelineData.value.flatMap((entry) => entry.data.flatMap((point) => point.y ?? []));
    if (!timestamps.length) return null;
    return { min: Math.min(...timestamps), max: Math.max(...timestamps) };
  });

  const hasTimelineData = computed(() => timelineExtent.value !== null);
  const rangeLabels = computed(() => {
    const range = timelineExtent.value;
    if (!range) {
      return null;
    }

    return {
      start: formatTimestamp(range.min),
      end: formatTimestamp(range.max),
    };
  });

  const chartOptions = computed(() => {
    const range = timelineExtent.value;
    const title = props.chartData?.[0]?.name ?? 'Timeline';

    return resolveChartOptions(ChartTypes.TimelineChart, {
      tooltip: timelineTooltip,
      title: {
        text: title,
        align: 'left',
        style: { color: '#aaa', fontSize: '12px' },
      },
      xaxis: {
        type: 'datetime',
        min: range?.min,
        max: range?.max,
        labels: { datetimeUTC: false },
      },
      chart: { sparkline: { enabled: false } },
    });
  });

  const legendColors = computed(() => getExposeBinaryColour(props.chartData?.[0]?.name ?? '')); 

  function buildSeries(expose: DeviceExposeBinaryMetrics, singleRow: boolean): TimelineChartEntry {
    const color = getExposeBinaryColour(expose.name);
    const merged = mergeRanges(expose.data, MERGE_TOLERANCE_MS);

    return {
      name: expose.name,
      data: merged.map((point) => ({
        x: singleRow ? expose.name : `${expose.name}: ${formatStateLabel(point.x)}`,
        y: point.y,
        state: point.x,
        label: expose.name,
        fillColor: point.x === 'true' ? color.on : color.off,
      })),
    };
  }

  function mergeRanges(points: BinaryPoint[], toleranceMs: number): BinaryPoint[] {
    if (!points?.length) return [];

    const sorted = [...points].sort((a, b) => (a.y?.[0] ?? 0) - (b.y?.[0] ?? 0));
    return sorted.reduce<BinaryPoint[]>((acc, current) => {
      const clone: BinaryPoint = {
        x: current.x,
        y: [current.y?.[0] ?? 0, current.y?.[1] ?? current.y?.[0] ?? 0],
      };
      const last = acc[acc.length - 1];

      if (last && last.x === clone.x && Math.abs((clone.y?.[0] ?? 0) - (last.y?.[1] ?? 0)) <= toleranceMs) {
        last.y = [last.y?.[0] ?? clone.y?.[0] ?? 0, Math.max(last.y?.[1] ?? 0, clone.y?.[1] ?? 0)];
        return acc;
      }

      acc.push(clone);
      return acc;
    }, []);
  }

  function formatStateLabel(value?: string) {
    if (value === 'true') return 'On';
    if (value === 'false') return 'Off';
    return value ?? '';
  }

  function formatTimestamp(value: number) {
    return new Date(value).toLocaleString('en-GB', {
      hour: '2-digit',
      minute: '2-digit',
      day: '2-digit',
      month: 'short',
    });
  }
</script>

<template>
  <div class="timeline-chart">
    <div v-if="hasTimelineData" class="timeline-chart__canvas">
      <div class="timeline-chart__controls">
        <span>{{ layoutLabel }}</span>
        <button type="button" class="toggle" @click="toggleRowLayout">
          {{ isSingleRow ? 'Show split view' : 'Show single view' }}
        </button>
      </div>
      <BaseChart width="100%" height="100%" :data="timelineData" :options="chartOptions" />
      <div v-if="rangeLabels" class="timeline-chart__range">
        <span>Start: {{ rangeLabels.start }}</span>
        <span>End: {{ rangeLabels.end }}</span>
      </div>
      <div class="timeline-chart__legend">
        <div class="legend-item">
          <span class="legend-swatch" :style="{ background: legendColors.on }"></span>
          <span>On</span>
        </div>
        <div class="legend-item">
          <span class="legend-swatch" :style="{ background: legendColors.off }"></span>
          <span>Off</span>
        </div>
      </div>
    </div>
    <div v-else class="timeline-chart__placeholder">No timeline data for this period.</div>
  </div>
</template>

<style scoped>
  .timeline-chart {
    width: 100%;
  }
  .timeline-chart__canvas {
    position: relative;
    width: 100%;
    height: 240px;
  }
  .timeline-chart__controls {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    font-size: 12px;
    color: #bbb;
    margin-bottom: 4px;
  }
  .timeline-chart__controls .toggle {
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 3px;
    color: #ddd;
    font-size: 12px;
    padding: 2px 8px;
    cursor: pointer;
  }
  .timeline-chart__controls .toggle:hover {
    border-color: rgba(255, 255, 255, 0.4);
  }
  .timeline-chart__legend {
    display: flex;
    gap: 12px;
    margin-top: 8px;
    font-size: 12px;
    color: #bbb;
  }
  .timeline-chart__range {
    display: flex;
    justify-content: space-between;
    margin-top: 6px;
    font-size: 12px;
    color: #aaa;
  }
  .legend-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .legend-swatch {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 3px;
  }
  .timeline-chart__placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 240px;
    border: 1px dashed rgba(255, 255, 255, 0.2);
    border-radius: 4px;
    color: #777;
    font-size: 13px;
  }
</style>
