import { ValueOf } from './types';

export type AppConfig = {
  devices: KeyyValuePair<DeviceSettings>;
  history: HistorySettingsType;
  logger: LoggerSettingsType;
  bridge: BridgeSettingsType;
};

export type DeviceSettingsMap = KeyyValuePair<DeviceSettings>;

export const TimeUnits: {
  UnitSeconds: 'seconds';
  UnitMinutes: 'minutes';
  UnitHours: 'hours';
  UnitDays: 'days';
};

export type TimeUnit = ValueOf<typeof TimeUnits>;

export type TimeInterval = {
  value: int;
  unit: TimeUnit;
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
  permitJoinEnabled: boolean;
  maxTimeAllowed: TimeInterval;
};

export type DeviceSettings = {
  id: string;
  disabled: boolean;
  metricsEnabled: boolean;
  rateLimit: number;
};
