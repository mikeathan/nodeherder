import { KeyValuePair, TimeInterval, ValueOf } from './types.type';

export type AppConfig = {
  hub: HubConfigType;
  bridge: BridgeSettingsType;
};

export type DeviceConfigMap = KeyValuePair<DeviceConfig>;
export type DashboardGroups = KeyValuePair<DashboardGroup>;

export type HubConfigType = {
  devices: DeviceSettings;
  history: HistorySettingsType;
  logger: LoggerSettingsType;
  mcp: MCPSettingsType;
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

export type MCPSettingsType = {
  enabled: boolean;
};

export type BridgeSettingsType = {
  permitJoin: boolean;
  maxTimeAllowed: TimeInterval;
};

export type DeviceSettings = {
  defaults: DeviceConfig;
  overrides: DeviceConfigMap;
};

export type DeviceConfig = {
  id: string;
  disabled: boolean;
  metricsEnabled: boolean;
  defaultDebounceByCategory: DeviceDebounce;
  debounceOverrides: DeviceDebounce;
};

export type DashboardGroup = {
  name: string;
  deviceGroup: KeyValuePair<DeviceGroup>;
};

export type DeviceGroup = {
  deviceId: string;
  exposes: string[];
};

export type MCPStatusType = {
  running: boolean;
  name: string;
  version: string;
  connectedClients: number;
};
