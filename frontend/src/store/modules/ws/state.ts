import { ConnectionStatusType } from '@/types/connection.type';

export interface WSClientState {
  socket: WebSocket | null;
  connectionStatus: ConnectionStatusType;
}
