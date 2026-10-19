<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
  import type { Chart } from 'chart.js';
  import type { HistoryChartOptions } from './options';
  type ChartPoint = { x: number; y: number | null };
  const props = defineProps<{
    type: 'line' | 'bar';
    height: number;
    endMarker?: boolean;
    options: HistoryChartOptions;
    series: { name: string; data: ChartPoint[] }[];
  }>();
  const canvas = ref<HTMLCanvasElement | null>(null);
  const ready = ref(false);
  const error = ref('');
  // A chart owns native listeners and observers, so keep its instance unproxied.
  let chart: Chart<'line' | 'bar', ChartPoint[]> | null = null;
  let disposed = false;
  const datasets = () =>
    props.series.map((series) => {
      const last = lastKnownIndex(series.data);
      return {
        label: series.name,
        data: series.data,
        // Mark the last actual sample, never the right edge of an unknown gap.
        pointRadius: (context: { dataIndex: number }) =>
          series.data.length === 1 || (props.endMarker && context.dataIndex === last) ? 3 : 0,
      };
    });
  function lastKnownIndex(data: ChartPoint[]) {
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
      chart = new Chart<'line' | 'bar', ChartPoint[]>(canvas.value, {
        type: props.type,
        data: { datasets: datasets() },
        options: props.options,
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
  <div class="history-canvas-wrap" :style="{ height: `${height}px` }">
    <p v-if="error" role="alert">{{ error }}</p>
    <canvas ref="canvas" class="nh-history-canvas" :data-ready="ready" aria-hidden="true" />
  </div>
</template>
<style scoped>
  .history-canvas-wrap {
    position: relative;
    width: 100%;
    min-width: 0;
  }
</style>
