<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
  import type { Chart } from 'chart.js';
  import type { HistoryChartOptions, HistoryChartPoint, HistoryChartSeries } from './options';
  import { historyInteractionPlugin, type HistoryInteraction } from './interactions';
  const props = defineProps<{
    type: 'line' | 'bar';
    height: number;
    endMarker?: boolean;
    options: HistoryChartOptions;
    interaction?: HistoryInteraction;
    series: HistoryChartSeries[];
  }>();
  const canvas = ref<HTMLCanvasElement | null>(null);
  const ready = ref(false);
  const error = ref('');
  // A chart owns native listeners and observers, so keep its instance unproxied.
  let chart: Chart<'line' | 'bar', HistoryChartPoint[]> | null = null;
  let disposed = false;
  const datasets = () =>
    props.series.map((series) => {
      const last = lastKnownIndex(series.data);
      return {
        label: series.name,
        data: series.data,
        ...(series.type ? { type: series.type } : {}),
        ...(series.color ? { backgroundColor: series.color, borderColor: series.color, pointBackgroundColor: series.color } : {}),
        ...(series.outline ? { borderColor: series.outline, borderWidth: 1, pointBorderColor: series.outline } : {}),
        ...(series.pointStyle ? { pointStyle: series.pointStyle, showLine: false, fill: false, pointBorderWidth: 2, pointHoverRadius: 6 } : {}),
        order: series.pointStyle ? 0 : 1,
        // Mark the last actual sample, never the right edge of an unknown gap.
        pointRadius: (context: { dataIndex: number }) =>
          series.pointStyle ? 5 : series.data.length === 1 || (props.endMarker && context.dataIndex === last) ? 3 : 0,
      };
    });
  function lastKnownIndex(data: HistoryChartPoint[]) {
    for (let i = data.length - 1; i >= 0; i--) if (data[i].y !== null) return i;
    return -1;
  }
  onMounted(async () => {
    try {
      const {
        Chart,
        LineController,
        BarController,
        LinearScale,
        LineElement,
        BarElement,
        PointElement,
        Tooltip,
        Filler,
      } = await import('chart.js');
      if (disposed || !canvas.value) return;
      Chart.register(
        LineController,
        BarController,
        LinearScale,
        LineElement,
        BarElement,
        PointElement,
        Tooltip,
        Filler
      );
      chart = new Chart<'line' | 'bar', HistoryChartPoint[]>(canvas.value, {
        type: props.type,
        data: { datasets: datasets() },
        options: props.options,
        plugins: props.interaction ? [historyInteractionPlugin(props.interaction)] : [],
      });
      ready.value = true;
    } catch {
      if (!disposed) error.value = 'The chart could not be displayed. Recorded samples are still available.';
    }
  });
  watch([() => props.options, () => props.series], () => {
    if (!chart) return;
    chart.options = props.options;
    chart.data.datasets = datasets();
    chart.update('none');
  });
  onBeforeUnmount(() => {
    disposed = true;
    chart?.destroy();
    chart = null;
  });
</script>
<template>
  <div class="history-canvas-wrap" :style="{ '--history-height': `${height}px` }">
    <p v-if="error" role="alert">{{ error }}</p>
    <canvas ref="canvas" class="nh-history-canvas" :data-ready="ready" :data-points="series.reduce((total, item) => total + item.data.length, 0)" aria-hidden="true" />
  </div>
</template>
<style scoped>
  .history-canvas-wrap {
    position: relative;
    width: 100%;
    min-width: 0;
    height: var(--nh-chart-height, var(--history-height));
  }
  .history-canvas-wrap canvas { touch-action: none; max-width: 100%; }
</style>
