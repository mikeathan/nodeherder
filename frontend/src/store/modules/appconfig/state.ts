import {
  AppConfig,
  DeviceSettingsMap,
} from '@/types/settings';

export interface AppConfigModuleState {
  deviceSettingsMap: DeviceSettingsMap;
  appConfig: AppConfig;
  initialized: boolean;
}
