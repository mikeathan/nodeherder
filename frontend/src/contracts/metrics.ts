import { ChartType, ChartTypes } from '@/types/chart.type';
import { DeviceMetrics, DeviceExposeMetrics, MetricsTypes } from '@/types/metrics.type';
import { KeyValuePair } from '@/types/types.type';
import { isBinaryChartableExpose } from '@/utils/chart.utils';

export function groupMetricsByType(metrics: DeviceMetrics): KeyValuePair<DeviceExposeMetrics[]> {
  if (!metrics) {
    return {};
  }
  return metrics.exposes.reduce(
    (grouped, expose) => {
      let chartType: ChartType = ChartTypes.NumericChart;
      if (isBinaryChartableExpose(expose.type, expose.name)) {
        chartType = ChartTypes.BinaryChart;
      } else if (expose.type === MetricsTypes.Enum) {
        return grouped; // Skip other enum types
      }
      grouped[chartType] = (grouped[chartType] || []).concat(expose);

      return grouped;
    },
    {} as KeyValuePair<DeviceExposeMetrics[]>
  );
}
