<
<script setup lang="ts">
  import { computed, ref } from 'vue';
  import type { PropType } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import { ChartTypes, TimelineChartEntry } from '@/types/chart.type';
  import { DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import { getExposeBinaryColour, resolveChartOptions } from '@/contracts/chart';

  type BinaryPoint = DeviceExposeBinaryMetrics['data'][number];
  type RowLayout = 'single' | 'split';
  const MERGE_TOLERANCE_MS = 30_000;

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

  // -----------------------------------------------------
  // Layout toggle (cheap)
  // -----------------------------------------------------
  const rowLayout = ref<RowLayout>(props.defaultRowLayout);
  const isSingle = computed(() => rowLayout.value === 'single');
  const layoutLabel = computed(() => (isSingle.value ? 'Single row' : 'Split rows'));

  function toggleRowLayout() {
    rowLayout.value = isSingle.value ? 'split' : 'single';
  }

  // -----------------------------------------------------
  // 1. Merge raw data ONCE
  // -----------------------------------------------------
  const mergedData = computed(() => {
    return (props.chartData ?? []).map((expose) => {
      const colour = getExposeBinaryColour(expose.name);
      const merged = mergeRanges(expose.data, MERGE_TOLERANCE_MS);

      return { name: expose.name, colour, merged };
    });
  });

  // -----------------------------------------------------
  // 2. Build chart-ready series (cheap mapping)
  // -----------------------------------------------------
  const timelineData = computed<TimelineChartEntry[]>(() => {
    const single = isSingle.value;

    return mergedData.value.map((item) => ({
      name: item.name,
      data: item.merged.map((p) => ({
        x: single ? item.name : `${item.name}: ${formatStateLabel(p.x)}`,
        y: p.y,
        state: p.x,
        label: item.name,
        fillColor: p.x === 'true' ? item.colour.on : item.colour.off,
      })),
    }));
  });

  // -----------------------------------------------------
  // 3. Compute extents (single pass)
  // -----------------------------------------------------
  const timelineExtent = computed(() => {
    let min = Infinity;
    let max = -Infinity;

    for (const e of mergedData.value) {
      for (const p of e.merged) {
        const [s, e2] = p.y;
        if (s < min) min = s;
        if (e2 > max) max = e2;
      }
    }

    return min === Infinity ? null : { min, max };
  });

  const hasTimelineData = computed(() => timelineExtent.value !== null);

  // Range labels
  const rangeLabels = computed(() => {
    if (!timelineExtent.value) return null;
    return {
      start: formatTimestamp(timelineExtent.value.min),
      end: formatTimestamp(timelineExtent.value.max),
    };
  });

  const chartOptions = computed(() => {
    const range = timelineExtent.value;
    const title = props.chartData?.[0]?.name ?? 'Timeline';

    return resolveChartOptions(ChartTypes.TimelineChart, {
      tooltip: { custom: timelineTooltip },
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

  // Legend colours
  const legendColors = computed(() => getExposeBinaryColour(props.chartData?.[0]?.name ?? ''));

  function mergeRanges(points: BinaryPoint[], tol: number): BinaryPoint[] {
    if (!points?.length) return [];

    const sorted = [...points].sort((a, b) => (a.y?.[0] ?? 0) - (b.y?.[0] ?? 0));

    const out: BinaryPoint[] = [];

    for (const c of sorted) {
      const s = c.y?.[0] ?? 0;
      const e = c.y?.[1] ?? s;

      if (!out.length) {
        out.push({ x: c.x, y: [s, e] });
        continue;
      }

      const last = out[out.length - 1];
      const lastEnd = last.y?.[1] ?? 0;

      if (last.x === c.x && Math.abs(s - lastEnd) <= tol) {
        last.y[1] = Math.max(lastEnd, e);
      } else {
        out.push({ x: c.x, y: [s, e] });
      }
    }

    return out;
  }

  function formatStateLabel(v?: string) {
    return v === 'true' ? 'On' : v === 'false' ? 'Off' : v ?? '';
  }

  function formatTimestamp(v: number) {
    return new Date(v).toLocaleString('en-GB', {
      hour: '2-digit',
      minute: '2-digit',
      day: '2-digit',
      month: 'short',
    });
  }

  function formatTime(v: number) {
    return new Date(v).toLocaleTimeString('en-GB', {
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  function timelineTooltip({ seriesIndex, dataPointIndex, w }: any) {
    const d = w.globals.initialSeries?.[seriesIndex]?.data?.[dataPointIndex];
    if (!d || !Array.isArray(d.y)) return '';

    const [s, e] = d.y;
    const dur = Math.max(0, e - s);
    const mins = Math.floor(dur / 60000);
    const hrs = Math.floor(mins / 60);
    const rem = mins % 60;

    const duration = hrs > 0 ? `${hrs}h ${rem}m` : `${mins}m`;

    return `
    <div style="padding:6px;font-size:12px;background:#1e1e1e;color:#fff;border-radius:4px">
      <b>${d.label}</b>: ${formatStateLabel(d.state)}<br/>
      ${formatTime(s)} → ${formatTime(e)}<br/>
      <small>${duration}</small>
    </div>
  `;
  }
</script>

<template>
  <div class="timeline-chart">
    <div v-if="hasTimelineData" class="timeline-chart__canvas">
      <div class="timeline-chart__controls">
        <span>{{ layoutLabel }}</span>
        <button type="button" class="toggle" @click="toggleRowLayout">
          {{ isSingle ? 'Show split view' : 'Show single view' }}
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
.timeline-chart { width: 100%; }
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
.toggle {
  background: transparent;
  border: 1px solid rgba(255,255,255,0.2);
  border-radius: 3px;
  color: #ddd;
  font-size: 12px;
  padding: 2px 8px;
  cursor: pointer;
}
.toggle:hover {
  border-color: rgba(255,255,255,0.4);
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
  width: 12px;
  height: 12px;
  border-radius: 3px;
}
.timeline-chart__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 240px;
  border: 1px dashed rgba(255,255,255,0.2);
  border-radius: 4px;
  color: #777;
  font-size: 13px;
}
</style>
