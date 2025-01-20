import {
  AppConfig,
  DeviceSettingsMap,
} from '@/types/settings.type';

export interface AppConfigModuleState {
  deviceSettingsMap: DeviceSettingsMap;
  appConfig: AppConfig;
  initialized: boolean;
}
