import { defineAsyncComponent } from 'vue';

type ChartType = string;
type ChartMap = { [key: ChartType]: any };

export const ChartComponents: ChartMap = {
  TimelineChart: defineAsyncComponent(
    () => import('../components/chart/TimelineChart.vue'),
  ),
  AreaChart: defineAsyncComponent(
    () => import('../components/chart/AreaChart.vue'),
  ),
};
