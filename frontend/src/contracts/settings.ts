import { DeviceSettings } from '@/types/settings';

export function createDeviceSettings(
  id: string,
  enabled: boolean = true,
  metricsEnabled: boolean = false,
  rateLimit: number = 50000
): DeviceSettings {
  return {
    id: id,
    disabled: !enabled,
    metricsEnabled: metricsEnabled,
    rateLimit: rateLimit,
  };
}
