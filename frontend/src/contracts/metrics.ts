import {
  DeviceMetrics,
  DeviceExposeMetrics,
} from '@/types/metrics';
import { KeyyValuePair } from '@/types/types';

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

TODO - fix;

export function groupMetricsByType(
  metrics: DeviceMetrics,
): KeyyValuePair<DeviceExposeMetrics[]> {
  let map: KeyyValuePair<DeviceExposeMetrics[]> = {};

  metrics.expose.forEach((expose) => {
    if (expose.type === 'int') {
      map['Timeline'].push(expose);
    } else {
      map['AreaChart'].push(expose);
    }
  });

  return map;
}
