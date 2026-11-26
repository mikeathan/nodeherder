/*
 * Connection notifications are produced from transport transitions, not from components.
 * One subscription owns dedupe and cleanup, so remounting a page cannot duplicate history
 * or leak a listener.
 */
import { onScopeDispose } from 'vue';
import { store } from '@/store';
import type { ConnectionStatusType } from '@/types/connection.type';
import { connectionTransition } from '@/domain/notifications';

/** A quick restart/reconnect is routine maintenance, not an incident worth an inbox row. */
export const OUTAGE_GRACE_MS = 30_000;

export function useNotificationProducers() {
  let previous = store.getters['ws/getConnectionStatus'] as ConnectionStatusType;
  let outageStartedAt: number | null = null;
  let lossNotified = false;
  let lossTimer: ReturnType<typeof setTimeout> | null = null;

  function clearOutage() {
    if (lossTimer !== null) clearTimeout(lossTimer);
    lossTimer = null;
    outageStartedAt = null;
    lossNotified = false;
  }

  const unsubscribe = store.subscribe((mutation) => {
    if (mutation.type !== 'ws/setConnectionStatus') return;
    const next = mutation.payload as ConnectionStatusType;
    if (next === previous) return;

    const transition = connectionTransition(previous, next);
    previous = next;

    if (transition === 'lost') {
      if (outageStartedAt !== null) return;
      outageStartedAt = Date.now();
      lossNotified = false;
      if (lossTimer !== null) clearTimeout(lossTimer);
      lossTimer = setTimeout(() => {
        lossTimer = null;
        if (outageStartedAt === null) return;
        lossNotified = true;
        void store.dispatch('notifications/notify', {
          severity: 'warn',
          category: 'connection',
          title: 'Hub connection lost',
          body: 'Live updates are paused. Values show the last snapshot until the link returns.',
          dedupeKey: 'hub-connection',
          action: { label: 'Open logs', to: '/consoleviewer' },
        });
      }, OUTAGE_GRACE_MS);
      return;
    }

    if (transition === 'restored' && outageStartedAt !== null) {
      const duration = Date.now() - outageStartedAt;
      const shouldResolve = lossNotified;
      clearOutage();
      if (shouldResolve) void store.dispatch('notifications/resolveConnection', duration);
    }
  });

  onScopeDispose(() => {
    unsubscribe();
    if (lossTimer !== null) clearTimeout(lossTimer);
    lossTimer = null;
  });
}
