import { ConnectionStatusType } from '@/types/connection.type';

export const connectionText = (status: ConnectionStatusType): string =>
  status === 'connected' ? 'Connected' : status === 'connecting' ? 'Reconnecting…' : 'Disconnected';
