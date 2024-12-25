import { ConnectionStatus } from '@/types/connection.type';
import { WsClientService } from './ws';

export interface WSClientState {
  ws: WsClientService;
  socket:WebSocket | null;
  connected: boolean;
  connectionStatus: ConnectionStatus;
}
