<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
  import type { Chart } from 'chart.js';
  import type { HistoryChartOptions } from './options';
  type ChartPoint = { x: number; y: number | null };
  const props = defineProps<{
    type: 'line' | 'bar';
    height: number;
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
    props.series.map((series) => ({
      label: series.name,
      data: series.data,
      // A lone reading still needs a visible mark when no line can be drawn.
      pointRadius: series.data.length === 1 ? 3 : 0,
    }));
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
