import { ChartType, ChartTypes } from '@/types/chart.type';
import {
  DeviceMetrics,
  DeviceExposeMetrics,
  MetricsTypes,
} from '@/types/metrics.type';
import { KeyValuePair } from '@/types/types';

export function groupMetricsByType(
  metrics: DeviceMetrics,
): KeyValuePair<DeviceExposeMetrics[]> {
  if (!metrics) {
    return {};
  }
  return metrics.expose.reduce((grouped, expose) => {
    let chartType: ChartType = ChartTypes.FloatChart;
    if (expose.type === MetricsTypes.Integer) {
      chartType = ChartTypes.BinaryChart;
    }
    grouped[chartType] = (grouped[chartType] || []).concat(
      expose,
    );

    return grouped;
  }, {} as KeyValuePair<DeviceExposeMetrics[]>);
}
