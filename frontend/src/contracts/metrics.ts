import { ChartType, ChartTypes } from '@/types/chart.type';
import {
  DeviceMetrics,
  DeviceExposeMetrics,
  MetricsTypes,
} from '@/types/metrics.type';
import { KeyyValuePair } from '@/types/types';

export function groupMetricsByType(
  metrics: DeviceMetrics,
): KeyyValuePair<DeviceExposeMetrics[]> {
  if (!metrics) {
    return {};
  }
  return metrics.expose.reduce((grouped, expose) => {
    let chartType: ChartType = ChartTypes.AreaChart;
    if (expose.type === MetricsTypes.Integer) {
      chartType = ChartTypes.TimelineChart;
    }
    grouped[chartType] = (grouped[chartType] || []).concat(
      expose,
    );

    return grouped;
  }, {} as KeyyValuePair<DeviceExposeMetrics[]>);
}
