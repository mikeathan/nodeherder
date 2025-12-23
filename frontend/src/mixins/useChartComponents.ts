import { defineAsyncComponent } from 'vue';

type ChartType = string;
type ChartMap = { [key: ChartType]: any };

export const ChartComponents: ChartMap = {
  NumericChart: defineAsyncComponent(() => import('../components/chart/NumericChart.vue')),
  BinaryChart: defineAsyncComponent(() => import('../components/chart/BinaryChart.vue')),
  AreaChart: defineAsyncComponent(() => import('../components/chart/types/AreaChart.vue')),
  RangeBarChart: defineAsyncComponent(() => import('../components/chart/types/RangeBarChart.vue')),
};

export const MiniChartComponents: ChartMap = {
  MiniNumericChart: defineAsyncComponent(() => import('../components/chart/mini/MiniNumericChart.vue')),
  MiniPercentChart: defineAsyncComponent(() => import('../components/chart/mini/MiniPercentChart.vue')),
  MiniRealtimeChart: defineAsyncComponent(() => import('../components/chart/mini/MiniRealtimeChart.vue')),
  MiniEnergyChart: defineAsyncComponent(() => import('../components/chart/mini/MiniEnergyChart.vue')),
  MiniBinaryChart: defineAsyncComponent(() => import('../components/chart/mini/MiniBinaryChart.vue')),
};
