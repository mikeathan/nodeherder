import { defineAsyncComponent } from 'vue';

type ChartType = string;
type ChartMap = { [key: ChartType]: any };

export const ChartComponents: ChartMap = {
  BinaryChart: defineAsyncComponent(
    () => import('../components/chart/BinaryChart.vue'),
  ),
  NumericChart: defineAsyncComponent(
    () => import('../components/chart/NumericChart.vue'),
  ),
  TimelineChart: defineAsyncComponent(
    () =>
      import('../components/chart/types/TimelineChart.vue'),
  ),
  AreaChart: defineAsyncComponent(
    () => import('../components/chart/types/AreaChart.vue'),
  ),
};
