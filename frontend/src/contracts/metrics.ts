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
  return metrics.exposes.reduce((grouped, expose) => {
    let chartType: ChartType = ChartTypes.NumericChart;
    if (expose.type === MetricsTypes.Binary) {
      chartType = ChartTypes.BinaryChart;
    }
    grouped[chartType] = (grouped[chartType] || []).concat(
      expose,
    );

    return grouped;
  }, {} as KeyValuePair<DeviceExposeMetrics[]>);
}
