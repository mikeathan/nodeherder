import { KeyValuePair, ValueOf } from './types.type';

export const TimeUnits = {
  UnitMilliseconds: 'milliseconds',
  UnitSeconds: 'seconds',
  UnitMinutes: 'minutes',
  UnitHours: 'hours',
  UnitDays: 'days',
} as const;

export type TimeUnit = ValueOf<typeof TimeUnits>;
export type TimeInterval = {
  value: number;
  unit: TimeUnit;
};

export type AppConfig = {
  hub: HubConfigType;
  bridge: BridgeSettingsType;
};

export type DeviceSettingsMap = KeyValuePair<DeviceSettings>;

export type HubConfigType = {
  devices: KeyValuePair<DeviceSettings>;
  history: HistorySettingsType;
  logger: LoggerSettingsType;
};

export type HistorySettingsPropsType = keyof HistorySettingsType;

export type HistorySettingsType = {
  sleepTimeout: TimeInterval;
  expireAt: TimeInterval;
};

export type LoggerSettingsTypePropsType = keyof LoggerSettingsType;

export type LoggerSettingsType = {
  enableRemoteLogger: boolean;
};

export type BridgeSettingsType = {
  permitJoin: boolean;
  maxTimeAllowed: TimeInterval;
};

export type DeviceSettings = {
  id: string;
  disabled: boolean;
  metricsEnabled: boolean;
  rateLimit: TimeInterval;
};
