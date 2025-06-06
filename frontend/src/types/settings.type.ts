import { KeyValuePair, TimeInterval, ValueOf } from './types.type';

export type AppConfig = {
  hub: HubConfigType;
  bridge: BridgeSettingsType;
};

export type DeviceSettingsMap = KeyValuePair<DeviceSettings>;
export type DashboardGroups = KeyValuePair<DashboardGroup>;

export type HubConfigType = {
  devices: KeyValuePair<DeviceSettings>;
  history: HistorySettingsType;
  logger: LoggerSettingsType;
  dashboardGroups: DashboardGroups;
};

export type HistorySettingsPropsType = keyof HistorySettingsType;

export type HistorySettingsType = {
  sleepTimeout: TimeInterval;
  expireAt: TimeInterval;
};

export type LoggerSettingsTypePropsType = keyof LoggerSettingsType;

export type DeviceDebounce = KeyValuePair<TimeInterval>;
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
  debounce: DeviceDebounce;
};

export type DashboardGroup = {
  name: string;
  deviceGroup: KeyValuePair<DeviceGroup>;
};


export type DeviceGroup = {
  deviceId: string;
  exposes: string[];
};
