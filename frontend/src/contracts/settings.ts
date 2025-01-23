import { AppConfig, DeviceSettings, DeviceSettingsMap } from '@/types/settings.type';

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
    rateLimit: { value: rateLimit, unit: 'milliseconds' },
  };
}

export function createAppconfig(): AppConfig {
  return {
    hub: {
      devices: {} as DeviceSettingsMap,
      history: {
        sleepTimeout: {
          value: 0,
          unit: 'days',
        },
        expireAt: {
          value: 0,
          unit: 'days',
        },
      },
      logger: {
        enableRemoteLogger: false,
      },
    },
    bridge: {
      permitJoin: false,
      maxTimeAllowed: { value: 0, unit: 'days' },
    },
  };
}
// }
// export type LoggerSettingsType = {
//   enableRemoteLogger: boolean;
// };

// export type BridgeSettingsType = {
//   permitJoin: boolean;
//   maxTimeAllowed: TimeInterval;
// };

// export type DeviceSettings = {
//   id: string;
//   disabled: boolean;
//   metricsEnabled: boolean;
//   rateLimit: TimeInterval;
// };
