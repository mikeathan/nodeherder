import { ConnectionStatus } from '@/types/connection.type';
import { WsClientService } from './ws';

export interface WSClientState {
  ws: WsClientService;
  connected: boolean;
  connectionStatus: ConnectionStatus;
}
