import type { DeviceMap } from '@/types/device';
import type { SpatialFrame } from '@/types/spatial';
import type { AppConfig, MCPStatusType } from '@/types/settings.type';
import type { HubLoadStatus } from '@/types/hub.type';
import type { KeyValuePair } from '@/types/types.type';

export interface HubStateModuleState {
  deviceMap: DeviceMap;
  /**
   * Latest validated spatial frame per device. Frames are snapshots of mapped slots, not
   * device state: they never replace expose values and are dropped with the device list.
   */
  spatialFrames: KeyValuePair<SpatialFrame>;
  appConfig: AppConfig;
  initialized: boolean;
  loadStatus: HubLoadStatus;
  mcpStatus: MCPStatusType | null;
}
