import { Device, DeviceMap } from '../../../types/device';
import { AppConfig } from '../../../types/settings.type';

export interface HubStateModuleState {
  deviceMap: DeviceMap;
  appConfig: AppConfig;
  initialized: boolean;
}
