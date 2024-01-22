import { WsClient } from "./ws";
export interface WSClientState {
  ws: WsClient;
  connected: boolean;
}
