import {  DeviceMap } from '../../../types/device';
import { AppConfig, MCPStatusType } from '../../../types/settings.type';

export interface HubStateModuleState {
  deviceMap: DeviceMap;
  appConfig: AppConfig;
  initialized: boolean;
  mcpStatus: MCPStatusType | null;
}
