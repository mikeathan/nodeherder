import { ValueOf } from './types.type';

export const ConnectionStatus = {
  connected: 'connected',
  disconnected: 'disconnected',
  connecting: 'connecting',
};
export type ConnectionStatusType = ValueOf<
  typeof ConnectionStatus
>;

export type ConnectionStateIcon = {
  icon: string;
  color: string;
};
