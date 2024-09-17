import { ValueOf } from './types';

export type AppConfig = {
  devices: KeyyValuePair<DeviceSettings>;
  history: HistorySettingsType;
};

export type DeviceSettingsMap =
  KeyyValuePair<DeviceSettings>;

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

export type HistorySettingsPropsType =
  keyof HistorySettingsType;

export type HistorySettingsType = {
  sleepTimeout: TimeInterval;
  expireAt: TimeInterval;
};

export type DeviceSettings = {
  id: string;
  disabled: boolean;
  metricsEnabled: boolean;
  rateLimit: number;
};
