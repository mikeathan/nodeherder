import { defineAsyncComponent } from 'vue';

type ChartType = string;
type ChartMap = { [key: ChartType]: any };

export const ChartComponents: ChartMap = {
  NumericChart: defineAsyncComponent(() => import('../components/chart/NumericChart.vue')),
  BinaryChart: defineAsyncComponent(() => import('../components/chart/BinaryChart.vue')),
  AreaChart: defineAsyncComponent(() => import('../components/chart/types/AreaChart.vue')),
  RangeBarChart: defineAsyncComponent(() => import('../components/chart/types/RangeBarChart.vue')),
};
