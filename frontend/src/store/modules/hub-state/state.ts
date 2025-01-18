import { Device, DeviceMap } from '../../../types/device';
import { AppConfig } from '../../../types/settings';

export interface HubStateModuleState {
  deviceMap: DeviceMap;
  appConfig: AppConfig;
  initialized: boolean;
}
