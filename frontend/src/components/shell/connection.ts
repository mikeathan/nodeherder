import type { ConnectionNotice, ConnectionStatusType } from '@/types/connection.type';
import type { SessionValidationStatus } from '@/types/auth.type';
import type { HubLoadStatus } from '@/types/hub.type';

const labels: Record<ConnectionStatusType, string> = {
  idle: 'Connecting…', connecting: 'Connecting…', connected: 'Connected',
  reconnecting: 'Reconnecting…', disconnected: 'Disconnected',
};
export const connectionText = (status: ConnectionStatusType): string => labels[status];

/** Startup is neutral. Only evidence of failure creates an actionable notice. */
export function connectionNotices(connection: ConnectionStatusType, validation: SessionValidationStatus, snapshot: HubLoadStatus): ConnectionNotice[] {
  const notices: ConnectionNotice[] = [];
  if (validation === 'unavailable') notices.push({
    kind: 'session', title: 'Session check unavailable.',
    text: 'Could not check your session with the hub. Your sign-in has been kept.', retry: 'Retry session check',
  });
  if (connection === 'reconnecting' || connection === 'disconnected') notices.push({
    kind: 'transport', title: connection === 'reconnecting' ? 'Reconnecting.' : 'Disconnected.',
    text: 'Values may be out of date and device controls are paused until the hub reconnects.', retry: 'Reconnect',
  });
  if (snapshot === 'failed') notices.push({
    kind: 'snapshot', title: 'Home data unavailable.',
    text: 'Could not load your home from the hub. You can retry without signing in again.', retry: 'Retry loading home',
  });
  return notices;
}
