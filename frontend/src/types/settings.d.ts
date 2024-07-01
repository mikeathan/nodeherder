export type AppConfig = {
  devices: KeyyValuePair<DeviceSettings>;
};

export type DeviceSettingsMap = KeyyValuePair<DeviceSettings>;


export type DeviceSettings = {
  id: string;
  disabled: boolean;
  metricsEnabled: boolean;
  rateLimit: number;
};
