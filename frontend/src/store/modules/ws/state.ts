import { WsClientService } from "./ws";

export interface WSClientState {
  ws: WsClientService;
  connected: boolean;
}
