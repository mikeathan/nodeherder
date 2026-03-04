import { AppConfig, DeviceConfig, DeviceDebounce, DeviceSettings } from '@/types/settings.type';
import { TimeInterval, TimeUnit } from '@/types/types.type';

export function isTimeInterval(obj: any): obj is TimeInterval {
  return obj && typeof obj === 'object' && typeof obj.unit === 'string' && typeof obj.value === 'number';
}

export function isDeviceDebounce(obj: any): obj is Record<string, TimeInterval> {
  if (!obj || typeof obj !== 'object' || Array.isArray(obj)) return false;

  return Object.values(obj).every(isTimeInterval);
}

export function createTimeinterval(value: number, unit: TimeUnit): TimeInterval {
  return { value: value, unit: unit };
}

export function createTimeIntervalFromSeconds(seconds: number): TimeInterval {
  return { value: seconds, unit: 'seconds' };
}
export function createTimeIntervalFromMinutes(minutes: number): TimeInterval {
  return { value: minutes, unit: 'minutes' };
}

export function createDeviceConfigOverride(id: string): DeviceConfig {
  return {
    id: id,
    disabled: false,
    metricsEnabled: false,
    defaultDebounceByCategory: {},
    debounceOverrides: {},
  } as DeviceConfig;
}

export function createDeviceConfig(): DeviceConfig {
  return {
    id: '',
    disabled: false,
    metricsEnabled: false,
    defaultDebounceByCategory: {},
    debounceOverrides: {},
  } as DeviceConfig;
}

export function createDeviceSettings(): DeviceSettings {
  return {
    defaults: createDeviceConfig(),
    overrides: {},
  };
}

export function createAppconfig(): AppConfig {
  return {
    hub: {
      mcp: { enabled: false },
      devices: createDeviceSettings(),
      history: {
        sleepTimeout: { value: 0, unit: 'days' },
        expireAt: { value: 0, unit: 'days' },
      },
      logger: {
        enableRemoteLogger: false,
        level: 'info',
      },
      dashboardGroups: {},
    },
    bridge: {
      permitJoin: false,
      maxTimeAllowed: { value: 0, unit: 'days' },
    },
  };
}
