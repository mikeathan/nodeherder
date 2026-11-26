/*
 * Transport transitions that deserve a durable notification. Kept pure so the mapping is
 * testable without a component or a socket.
 */
import type { ConnectionStatusType } from '@/types/connection.type';

export type ConnectionTransition = 'lost' | 'restored' | null;

/** Maps one status change to a notification event; callers own the observed-loss guard. */
export function connectionTransition(previous: ConnectionStatusType, next: ConnectionStatusType): ConnectionTransition {
  if (previous === 'connected' && (next === 'reconnecting' || next === 'disconnected')) return 'lost';
  if (next === 'connected' && (previous === 'reconnecting' || previous === 'disconnected')) return 'restored';
  return null;
}
