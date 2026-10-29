import type { ConnectionInstance } from '@/types/integrations.type';
import type { WasteSchedulePresentation } from '@/domain/waste-schedule';
/** Project lifecycle/freshness precedence for already-presented resources, without transport or mutation. */
export function connectionStatus(
  instance: ConnectionInstance,
  resources: readonly WasteSchedulePresentation[],
  isWaste: boolean
): string {
  if (!instance.enabled) return 'Disabled';
  if (instance.last_error) return instance.last_error.message;
  if (instance.operation_id) return 'Checking for updates';
  if (resources.some((resource) => resource.freshness === 'stale')) return 'Needs refreshing';
  if (resources.length) return resources[0].statusLabel;
  return isWaste ? 'Waiting for collection dates' : 'Waiting for information';
}
