export type AppConfig = {
  devices: KeyyValuePair<DeviceSettings>;
  history: HistorySettings;
};

export type DeviceSettingsMap =
  KeyyValuePair<DeviceSettings>;

export type HistorySettings = {
  sleepTimeout: number;
  expireAt: number;
};

export type DeviceSettings = {
  id: string;
  disabled: boolean;
  metricsEnabled: boolean;
  rateLimit: number;
};
