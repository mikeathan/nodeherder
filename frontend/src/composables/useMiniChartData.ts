import { ref, watch, onUnmounted } from 'vue';
import { store } from '@/store/index';
import {
  DeviceMetrics,
  DeviceExposeMetrics,
  MetricsTypes,
  NumericDataPoint,
  BinaryDataPoint,
} from '@/types/metrics.type';
import { toUnix } from '@/utils/date.utils';

export interface MiniChartData {
  isLoading: boolean;
  hasData: boolean;
  data: NumericDataPoint[] | BinaryDataPoint[];
  type: string;
}

/**
 * Composable for fetching and managing mini chart data for a specific device expose
 * @param deviceId - The device ID
 * @param exposeName - The expose name (e.g., 'temperature', 'humidity')
 * @param duration - Duration in hours to fetch (default 24 hours)
 * @param enabled - Whether to fetch data (default true)
 */
export function useMiniChartData(deviceId: string, exposeName: string, duration: number = 24, enabled: boolean = true) {
  const chartData = ref<MiniChartData>({
    isLoading: false,
    hasData: false,
    data: [],
    type: MetricsTypes.Numeric,
  });

  const isLiveUpdates = ref<boolean>(false);
  let unsubscribe: (() => void) | null = null;

  const fetchData = () => {
    if (!deviceId || !exposeName) return;

    chartData.value.isLoading = true;

    const now = new Date();
    const from = new Date(now.getTime() - duration * 60 * 60 * 1000);

    const request = {
      id: deviceId,
      expose: exposeName,
      from: toUnix(from),
      to: toUnix(now),
    };

    // Clean up existing subscription before creating new one
    if (unsubscribe) {
      unsubscribe();
      unsubscribe = null;
    }

    // Subscribe to store changes to get the data
    let dataReceived = false;
    unsubscribe = store.watch(
      (state: any) => state.metrics.deviceMetricsQueryMap[deviceId],
      (metrics: DeviceMetrics) => {
        if (!metrics) return;

        const exposeData = metrics.exposes.find((exp: DeviceExposeMetrics) => exp.name === exposeName);

        if (exposeData) {
          chartData.value.type = exposeData.type;
          chartData.value.data = (exposeData as any).data || [];
          chartData.value.hasData = chartData.value.data.length > 0;
          chartData.value.isLoading = false;

          // If live updates are not enabled and we got the data, unsubscribe
          if (!isLiveUpdates.value && !dataReceived) {
            dataReceived = true;
            if (unsubscribe) {
              unsubscribe();
              unsubscribe = null;
            }
          }
        }
      },
      { deep: true }
    );

    // Dispatch metrics query
    store.dispatch('metrics/query', request);
  };

  const toggleLiveUpdates = () => {
    isLiveUpdates.value = !isLiveUpdates.value;

    if (isLiveUpdates.value) {
      // Start watching for updates
      fetchData();
    } else {
      // Stop watching for updates
      if (unsubscribe) {
        unsubscribe();
        unsubscribe = null;
      }
    }
  };

  // Auto-fetch if enabled
  if (enabled) {
    fetchData();
  }

  // Cleanup subscription on unmount
  onUnmounted(() => {
    if (unsubscribe) {
      unsubscribe();
    }
  });

  return {
    chartData,
    refetch: fetchData,
    isLiveUpdates,
    toggleLiveUpdates,
  };
}
