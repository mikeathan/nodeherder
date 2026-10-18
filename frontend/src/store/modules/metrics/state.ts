import { DeviceMetricsMap } from '@/types/metrics.type';
import type { HistoryEntry } from '@/domain/history/types';

export interface MetricsModuleState {
  epoch: number;
  history: Record<string, HistoryEntry>;
  deviceMetricsQueryMap: DeviceMetricsMap;
}
