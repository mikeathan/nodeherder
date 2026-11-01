import type { SpatialMapping, SpatialWorkspace } from './spatial';
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
  assistant: AssistantSettingsType;
  dashboardGroups: DashboardGroups;
};

export type HistorySettingsPropsType = keyof HistorySettingsType;

export type HistorySettingsType = {
  sleepTimeout: TimeInterval;
  expireAt: TimeInterval;
};

export type AssistantSettingsPropsType = keyof AssistantSettingsType;

export type LoggerSettingsTypePropsType = keyof LoggerSettingsType;

export type DeviceDebounce = KeyValuePair<TimeInterval>;
export type LoggerSettingsType = {
  enableRemoteLogger: boolean;
  level: string;
};

export type MCPSettingsType = {
  enabled: boolean;
};

export type AssistantSettingsType = {
  url: string;
};

export type BridgeSettingsType = {
  permitJoin: boolean;
  maxTimeAllowed: TimeInterval;
};

export type DeviceSettings = {
  defaults: DeviceConfig;
  overrides: DeviceConfigMap;
};

/** Per-device spatial view settings persisted with the hub configuration. */
export type SpatialSettings = {
  /** Optional visual room/zone layout; absent for legacy records. */
  workspace?: SpatialWorkspace;
  mappings: SpatialMapping[];
  /** Display range in millimetres. */
  scale: number;
  /** View rotation in degrees. */
  orientation: number;
};

export type DeviceConfig = {
  spatial?: SpatialSettings;
  liveIntervalByCategory?: DeviceDebounce;
  liveIntervalOverrides?: DeviceDebounce;
  reportingThresholds?: Record<string, number>;
  metricOverrides?: Record<string, boolean>;
  commandConfirmationTimeoutMs?: number;
  id: string;
  disabled: boolean;
  metricsEnabled: boolean;
  defaultDebounceByCategory: DeviceDebounce;
  debounceOverrides: DeviceDebounce;
};

export type DashboardGroup = {
  name: string;
  deviceGroup: KeyValuePair<DeviceGroup>;
  /** Display position (spec 007 FR-09). Absent on groups saved before ordering existed. */
  order?: number;
};

export type DeviceGroup = {
  deviceId: string;
  exposes: string[];
};

export type MCPStatusType = {
  running: boolean;
  enabled: boolean;
  name: string;
  version: string;
  connectedClients: number;
};

