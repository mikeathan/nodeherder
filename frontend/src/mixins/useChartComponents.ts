import { defineAsyncComponent } from 'vue';

type ChartType = string;
type ChartMap = { [key: ChartType]: any };

export const ChartComponents: ChartMap = {
  NumericChart: defineAsyncComponent(() => import('../components/chart/NumericChart.vue')),
  BinaryChart: defineAsyncComponent(() => import('../components/chart/BinaryChart.vue')),
  DynamicBinaryChart: defineAsyncComponent(() => import('../components/chart/DynamicBinaryChart.vue')),
  AreaChart: defineAsyncComponent(() => import('../components/chart/types/AreaChart.vue')),
};

export const MiniChartComponents: ChartMap = {
  MiniNumericChart: defineAsyncComponent(() => import('../components/chart/mini/MiniNumericChart.vue')),
  MiniPercentChart: defineAsyncComponent(() => import('../components/chart/mini/MiniPercentChart.vue')),
  MiniRealtimeChart: defineAsyncComponent(() => import('../components/chart/mini/MiniRealtimeChart.vue')),
  MiniEnergyChart: defineAsyncComponent(() => import('../components/chart/mini/MiniEnergyChart.vue')),
  MiniDynamicBinaryChart: defineAsyncComponent(() => import('../components/chart/mini/MiniDynamicBinaryChart.vue')),
};
