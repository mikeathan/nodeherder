import { defineAsyncComponent } from 'vue';

type ChartType = string;
type ChartMap = { [key: ChartType]: any };

export const ChartComponents: ChartMap = {
  TimeRangeChart: defineAsyncComponent(
    () => import('../components/chart/TimeRangeChart.vue'),
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
