import type { DeviceConfigSaveResult } from '@/contracts/device-settings';
import type { ConnectionStatusType } from '@/types/connection.type';

export interface WSClientState {
  deviceConfigSaveResult: DeviceConfigSaveResult | null;
  socket: WebSocket | null;
  connectionStatus: ConnectionStatusType;
}
