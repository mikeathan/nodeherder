import type { ValueOf } from './types.type';

export const ConnectionStatus = {
  idle: 'idle',
  connected: 'connected',
  disconnected: 'disconnected',
  connecting: 'connecting',
  reconnecting: 'reconnecting',
} as const;
export type ConnectionStatusType = ValueOf<typeof ConnectionStatus>;

export type ConnectionStateIcon = {
  icon: string;
  color: string;
};

export type ConnectionNotice = {
  kind: 'session' | 'transport' | 'snapshot';
  title: string;
  text: string;
  retry: string;
};
