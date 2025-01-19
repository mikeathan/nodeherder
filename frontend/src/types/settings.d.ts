import { Device } from './device';
import { ValueOf } from './types';

export type AppConfig = {
  hub: HubConfigType;
  bridge: BridgeSettingsType;
};

export type DeviceSettingsMap = KeyyValuePair<DeviceSettings>;

export type HubConfigType = {
  devices: KeyyValuePair<DeviceSettings>;
  history: HistorySettingsType;
  logger: LoggerSettingsType;
};

export const TimeUnits: {
  UnitSeconds: 'seconds';
  UnitMinutes: 'minutes';
  UnitHours: 'hours';
  UnitDays: 'days';
};

export type HubState = {
  config: AppConfig;
  devices: Device[];
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
  permitJoin: boolean;
  maxTimeAllowed: TimeInterval;
};

export type DeviceSettings = {
  id: string;
  disabled: boolean;
  metricsEnabled: boolean;
  rateLimit: number;
};
