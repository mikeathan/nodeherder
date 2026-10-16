import type { DeviceMap } from '@/types/device';
import type { AppConfig, MCPStatusType } from '@/types/settings.type';
import type { HubLoadStatus } from '@/types/hub.type';

export interface HubStateModuleState {
  deviceMap: DeviceMap;
  appConfig: AppConfig;
  initialized: boolean;
  loadStatus: HubLoadStatus;
  mcpStatus: MCPStatusType | null;
}
