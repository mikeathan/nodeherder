import AreaChart from '@/components/chart/AreaChart.vue';
import { ChartType, ChartTypes } from '@/types/chart.type';
import {
  DeviceMetrics,
  DeviceExposeMetrics,
  MetricsTypes,
} from '@/types/metrics';
import { GenericMap, KeyyValuePair } from '@/types/types';

// export function groupMetricsByType(
//   metrics: DeviceMetrics,
// ): KeyyValuePair<DeviceExposeMetrics[]> {
//   if (!metrics) {
//     return {};
//   }
//   return metrics.expose.reduce((grouped, expose) => {
//     grouped[expose.type] = (
//       grouped[expose.type] || []
//     ).concat(expose);
//     return grouped;
//   }, {} as KeyyValuePair<DeviceExposeMetrics[]>);
// }


export function groupMetricsByType(
  metrics: DeviceMetrics,
): KeyyValuePair<DeviceExposeMetrics[]> {
  let map: GenericMap<ChartType, DeviceExposeMetrics[]> = {};

  metrics.expose.forEach((expose) => {
    if (expose.type === MetricsTypes.Integer) {
        
      map[ChartTypes.TimelineChart] ?? [].push(expose);
    } else {
      map[ChartTypes.AreaChart].push(expose);
    }
  });

  return map;
}
